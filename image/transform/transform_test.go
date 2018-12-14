package transform_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/crypto"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/image"
	imageTest "github.com/tidepool-org/platform/image/test"
	imageTransform "github.com/tidepool-org/platform/image/transform"
	imageTransformTest "github.com/tidepool-org/platform/image/transform/test"
	"github.com/tidepool-org/platform/pointer"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
)

var _ = Describe("Transform", func() {
	Context("Transform", func() {
		var rendition *image.Rendition

		BeforeEach(func() {
			rendition = imageTest.RandomRendition()
		})

		Context("NewTransform", func() {
			It("returns successfully with default values", func() {
				Expect(imageTransform.NewTransform()).To(Equal(&imageTransform.Transform{}))
			})
		})

		Context("NewTransformWithRendition", func() {
			It("returns an error if the rendition is missing", func() {
				rendition = nil
				transform, err := imageTransform.NewTransformWithRendition(rendition)
				Expect(err).To(MatchError("rendition is missing"))
				Expect(transform).To(BeNil())
			})

			It("returns an error if the rendition width is missing", func() {
				rendition.Width = nil
				transform, err := imageTransform.NewTransformWithRendition(rendition)
				Expect(err).To(MatchError("rendition width is missing"))
				Expect(transform).To(BeNil())
			})

			It("returns an error if the rendition height is missing", func() {
				rendition.Height = nil
				transform, err := imageTransform.NewTransformWithRendition(rendition)
				Expect(err).To(MatchError("rendition height is missing"))
				Expect(transform).To(BeNil())
			})

			It("returns successfully with expected values", func() {
				transform, err := imageTransform.NewTransformWithRendition(rendition)
				Expect(err).ToNot(HaveOccurred())
				Expect(transform).ToNot(BeNil())
				Expect(transform.Rendition).To(Equal(*rendition))
				Expect(transform.ContentWidth).To(Equal(*rendition.Width))
				Expect(transform.ContentHeight).To(Equal(*rendition.Height))
				Expect(transform.Resize).To(BeFalse())
				Expect(transform.Crop).To(BeFalse())
			})
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *imageTransform.Transform), expectedErrors ...error) {
					datum := imageTransformTest.RandomTransform()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New().Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *imageTransform.Transform) {},
				),
				Entry("rendition invalid",
					func(datum *imageTransform.Transform) {
						datum.Rendition.MediaType = nil
						datum.Rendition.Quality = nil
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rendition/mediaType"),
				),
				Entry("rendition valid",
					func(datum *imageTransform.Transform) { datum.Rendition = *imageTest.RandomRendition() },
				),
				Entry("width out of range (lower)",
					func(datum *imageTransform.Transform) { datum.ContentWidth = 0 },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 10000), "/width"),
				),
				Entry("width in range (lower)",
					func(datum *imageTransform.Transform) { datum.ContentWidth = 1 },
				),
				Entry("width in range (upper)",
					func(datum *imageTransform.Transform) { datum.ContentWidth = 10000 },
				),
				Entry("width out of range (upper)",
					func(datum *imageTransform.Transform) { datum.ContentWidth = 10001 },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10001, 1, 10000), "/width"),
				),
				Entry("height out of range (lower)",
					func(datum *imageTransform.Transform) { datum.ContentHeight = 0 },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 10000), "/height"),
				),
				Entry("height in range (lower)",
					func(datum *imageTransform.Transform) { datum.ContentHeight = 1 },
				),
				Entry("height in range (upper)",
					func(datum *imageTransform.Transform) { datum.ContentHeight = 10000 },
				),
				Entry("height out of range (upper)",
					func(datum *imageTransform.Transform) { datum.ContentHeight = 10001 },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(10001, 1, 10000), "/height"),
				),
				Entry("resize false; crop false",
					func(datum *imageTransform.Transform) {
						datum.Resize = false
						datum.Crop = false
					},
				),
				Entry("resize false; crop true",
					func(datum *imageTransform.Transform) {
						datum.Resize = false
						datum.Crop = true
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueBoolNotFalse(), "/crop"),
				),
				Entry("resize true; crop false",
					func(datum *imageTransform.Transform) {
						datum.Resize = true
						datum.Crop = false
					},
				),
				Entry("resize true; crop true",
					func(datum *imageTransform.Transform) {
						datum.Resize = true
						datum.Crop = true
					},
				),
				Entry("multiple errors",
					func(datum *imageTransform.Transform) {
						datum.Rendition.MediaType = nil
						datum.Rendition.Quality = nil
						datum.ContentWidth = 0
						datum.ContentHeight = 0
						datum.Resize = false
						datum.Crop = true
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotExists(), "/rendition/mediaType"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 10000), "/width"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(0, 1, 10000), "/height"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueBoolNotFalse(), "/crop"),
				),
			)
		})

		Context("with new transform", func() {
			var datum *imageTransform.Transform
			var original *imageTransform.Transform

			BeforeEach(func() {
				datum = imageTransformTest.RandomTransform()
			})

			JustBeforeEach(func() {
				original = imageTransformTest.CloneTransform(datum)
			})

			Context("with aspect ratio", func() {
				var aspectRatio float64

				BeforeEach(func() {
					aspectRatio = test.RandomFloat64FromRange(0.01, 100.0)
				})

				Context("ConstrainContentWidth", func() {
					It("constrains the content width based upon the height", func() {
						datum.ConstrainContentWidth(aspectRatio)
						Expect(datum.ContentWidth).To(Equal(int(math.Round(float64(original.ContentHeight) * aspectRatio))))
						Expect(datum.ContentHeight).To(Equal(original.ContentHeight))
						Expect(datum.Rendition.Width).To(Equal(original.Rendition.Width))
					})
				})

				Context("ConstrainContentHeight", func() {
					It("constrains the content height based upon the width", func() {
						datum.ConstrainContentHeight(aspectRatio)
						Expect(datum.ContentWidth).To(Equal(original.ContentWidth))
						Expect(datum.ContentHeight).To(Equal(int(math.Round(float64(original.ContentWidth) / aspectRatio))))
						Expect(datum.Rendition.Height).To(Equal(original.Rendition.Height))
					})
				})

				Context("ConstrainWidth", func() {
					It("constrains the width based upon the height", func() {
						datum.ConstrainWidth(aspectRatio)
						Expect(datum.ContentWidth).To(Equal(int(math.Round(float64(original.ContentHeight) * aspectRatio))))
						Expect(datum.ContentHeight).To(Equal(original.ContentHeight))
						Expect(datum.Rendition.Width).To(Equal(&datum.ContentWidth))
					})
				})

				Context("ConstrainHeight", func() {
					It("constrains the height based upon the width", func() {
						datum.ConstrainHeight(aspectRatio)
						Expect(datum.ContentWidth).To(Equal(original.ContentWidth))
						Expect(datum.ContentHeight).To(Equal(int(math.Round(float64(original.ContentWidth) / aspectRatio))))
						Expect(datum.Rendition.Height).To(Equal(&datum.ContentHeight))
					})
				})
			})

			Context("Reset", func() {
				var contentAttributes *image.ContentAttributes

				BeforeEach(func() {
					contentAttributes = imageTest.RandomContentAttributes()
				})

				It("returns an error if the content attributes is missing", func() {
					contentAttributes = nil
					Expect(datum.Reset(contentAttributes)).To(MatchError("content attributes is missing"))
				})

				It("returns an error if the content attributes width is missing", func() {
					contentAttributes.Width = nil
					Expect(datum.Reset(contentAttributes)).To(MatchError("content attributes width is missing"))
				})

				It("returns an error if the content attributes height is missing", func() {
					contentAttributes.Height = nil
					Expect(datum.Reset(contentAttributes)).To(MatchError("content attributes height is missing"))
				})

				When("the rendition media type is image/jpeg", func() {
					BeforeEach(func() {
						datum.Rendition.MediaType = pointer.FromString(image.MediaTypeImageJPEG)
					})

					It("sets expected values from content attributes and returns successfully", func() {
						Expect(datum.Reset(contentAttributes)).To(Succeed())
						Expect(datum.Rendition.MediaType).To(Equal(original.Rendition.MediaType))
						Expect(datum.Rendition.Width).To(Equal(contentAttributes.Width))
						Expect(datum.Rendition.Height).To(Equal(contentAttributes.Height))
						Expect(datum.Rendition.Mode).To(Equal(pointer.FromString(image.ModeScale)))
						Expect(datum.Rendition.Background).To(Equal(original.Rendition.Background))
						Expect(datum.Rendition.Quality).To(Equal(pointer.FromInt(image.QualityDefault)))
						Expect(datum.ContentWidth).To(Equal(*contentAttributes.Width))
						Expect(datum.ContentHeight).To(Equal(*contentAttributes.Height))
						Expect(datum.Resize).To(BeFalse())
						Expect(datum.Crop).To(BeFalse())
					})
				})

				When("the rendition media type is image/png", func() {
					BeforeEach(func() {
						datum.Rendition.MediaType = pointer.FromString(image.MediaTypeImagePNG)
						datum.Rendition.Quality = nil
					})

					It("sets expected values from content attributes and returns successfully", func() {
						Expect(datum.Reset(contentAttributes)).To(Succeed())
						Expect(datum.Rendition.MediaType).To(Equal(original.Rendition.MediaType))
						Expect(datum.Rendition.Width).To(Equal(contentAttributes.Width))
						Expect(datum.Rendition.Height).To(Equal(contentAttributes.Height))
						Expect(datum.Rendition.Mode).To(Equal(pointer.FromString(image.ModeScale)))
						Expect(datum.Rendition.Background).To(Equal(original.Rendition.Background))
						Expect(datum.Rendition.Quality).To(BeNil())
						Expect(datum.ContentWidth).To(Equal(*contentAttributes.Width))
						Expect(datum.ContentHeight).To(Equal(*contentAttributes.Height))
						Expect(datum.Resize).To(BeFalse())
						Expect(datum.Crop).To(BeFalse())
					})
				})
			})
		})
	})

	Context("Transformer", func() {
		Context("NewTransformer", func() {
			It("returns successfully", func() {
				Expect(imageTransform.NewTransformer()).ToNot(BeNil())
			})
		})

		Context("with new transformer", func() {
			var transformer *imageTransform.TransformerImpl

			BeforeEach(func() {
				transformer = imageTransform.NewTransformer()
			})

			Context("CalculateTransform", func() {
				var contentAttributes *image.ContentAttributes
				var rendition *image.Rendition

				BeforeEach(func() {
					contentAttributes = imageTest.RandomContentAttributes()
					rendition = imageTest.RandomRendition()
				})

				It("returns an error if the content attributes is missing", func() {
					contentAttributes = nil
					transform, err := transformer.CalculateTransform(contentAttributes, rendition)
					Expect(err).To(MatchError("content attributes is missing"))
					Expect(transform).To(BeNil())
				})

				It("returns an error if the content attributes is invalid", func() {
					contentAttributes.DigestMD5 = nil
					transform, err := transformer.CalculateTransform(contentAttributes, rendition)
					Expect(err).To(MatchError("content attributes is invalid; value does not exist"))
					Expect(transform).To(BeNil())
				})

				It("returns an error if the rendition is missing", func() {
					rendition = nil
					transform, err := transformer.CalculateTransform(contentAttributes, rendition)
					Expect(err).To(MatchError("rendition is missing"))
					Expect(transform).To(BeNil())
				})

				It("returns an error if the rendition is invalid", func() {
					rendition.MediaType = nil
					rendition.Quality = nil
					transform, err := transformer.CalculateTransform(contentAttributes, rendition)
					Expect(err).To(MatchError("rendition is invalid; value does not exist"))
					Expect(transform).To(BeNil())
				})
			})

			Context("TransformContent", func() {
				var reader io.Reader
				var transform *imageTransform.Transform

				BeforeEach(func() {
					reader = bytes.NewReader(imageTest.RandomContentBytes())
					transform = imageTransformTest.RandomTransform()
				})

				It("returns an error if the reader is missing", func() {
					reader = nil
					transform, err := transformer.TransformContent(reader, transform)
					Expect(err).To(MatchError("reader is missing"))
					Expect(transform).To(BeNil())
				})

				It("returns an error if the transform is missing", func() {
					transform = nil
					transform, err := transformer.TransformContent(reader, transform)
					Expect(err).To(MatchError("transform is missing"))
					Expect(transform).To(BeNil())
				})

				It("returns an error if the transform is invalid", func() {
					transform.ContentWidth = 0
					transform, err := transformer.TransformContent(reader, transform)
					Expect(err).To(MatchError("transform is invalid; value 0 is not between 1 and 10000"))
					Expect(transform).To(BeNil())
				})

				It("returns an error if it is unable to decode the content", func() {
					reader = bytes.NewReader(test.RandomBytes())
					transform, err := transformer.TransformContent(reader, transform)
					Expect(err).To(MatchError("unable to decode content; image: unknown format"))
					Expect(transform).To(BeNil())
				})
			})

			Context("Transform", func() {
				DescribeTable("properly transforms the fixture",
					func(sourceBase string, renditionAsString string) {
						By(fmt.Sprintf("transforming %q with rendition %q", sourceBase, renditionAsString), func() {
							sourcePath := filepath.Join(sourcesDirectory, sourceBase)

							sourceBytes, err := ioutil.ReadFile(sourcePath)
							Expect(err).ToNot(HaveOccurred(), "failure reading bytes from source file")

							sourceMediaType := mime.TypeByExtension(filepath.Ext(sourceBase))
							Expect(image.IsValidMediaType(sourceMediaType)).To(BeTrue(), "unexpected media type for source file")

							matches := sourceBaseRegexp.FindStringSubmatch(sourceBase)
							Expect(matches).To(HaveLen(3), "failure parsing source base")
							sourceWidth, err := strconv.ParseInt(matches[1], 10, 0)
							Expect(err).ToNot(HaveOccurred(), "failure parsing source width from source base")
							sourceHeight, err := strconv.ParseInt(matches[2], 10, 0)
							Expect(err).ToNot(HaveOccurred(), "failure parsing source height from source base")

							sourceFileInfo, err := os.Stat(sourcePath)
							Expect(err).ToNot(HaveOccurred(), "failure stating source file")

							sourceContentAttributes := image.NewContentAttributes()
							sourceContentAttributes.DigestMD5 = pointer.FromString(crypto.Base64EncodedMD5Hash(sourceBytes))
							sourceContentAttributes.MediaType = pointer.FromString(sourceMediaType)
							sourceContentAttributes.Width = pointer.FromInt(int(sourceWidth))
							sourceContentAttributes.Height = pointer.FromInt(int(sourceHeight))
							sourceContentAttributes.Size = pointer.FromInt(int(sourceFileInfo.Size()))
							sourceContentAttributes.CreatedTime = pointer.FromTime(time.Now())

							renditionBytes, err := ioutil.ReadFile(filepath.Join(renditionsDirectory, sourceBase, renditionAsString))
							Expect(err).ToNot(HaveOccurred(), "failure reading bytes from rendition file")

							rendition, err := image.ParseRenditionFromString(renditionAsString)
							Expect(err).ToNot(HaveOccurred(), "failure parsing rendition string")

							transform, err := transformer.CalculateTransform(sourceContentAttributes, rendition)
							Expect(err).ToNot(HaveOccurred(), "failure calculating transform")

							transformReader, err := transformer.TransformContent(bytes.NewReader(sourceBytes), transform)
							Expect(err).ToNot(HaveOccurred(), "failure transforming content")
							defer func() {
								Expect(transformReader.Close()).To(Succeed(), "failure closing transform reader")
							}()

							transformBytes, err := ioutil.ReadAll(transformReader)
							Expect(err).ToNot(HaveOccurred(), "failure reading bytes from transform reader")

							Expect(transformBytes).To(Equal(renditionBytes), "transformed bytes do not match expected rendition bytes")
						})
					},
					Entry("", "astronaut_180_240.jpeg", "w=100_h=150_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=100_h=150_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=100_h=150_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=100_h=150_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=100_h=150_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=100_h=150_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=100_h=150_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=100_h=150_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=100_h=150_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=100_h=150_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=100_h=150_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=100_h=150_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=100_h=150_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=100_h=150_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=100_h=150_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=100_h=150_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=240_h=240_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=240_h=240_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=240_h=240_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=240_h=240_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=240_h=240_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=240_h=240_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=240_h=240_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=240_h=240_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=240_h=240_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=240_h=240_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=240_h=240_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=240_h=240_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=240_h=240_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=240_h=240_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=240_h=240_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=240_h=240_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=250_h=250_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=250_h=250_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=250_h=250_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=250_h=250_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=250_h=250_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=250_h=250_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=250_h=250_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=250_h=250_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=250_h=250_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=250_h=250_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=250_h=250_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=250_h=250_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=250_h=250_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=250_h=250_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=250_h=250_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=250_h=250_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=300_h=200_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=300_h=200_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=300_h=200_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=300_h=200_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=300_h=200_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=300_h=200_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=300_h=200_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=300_h=200_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=300_h=200_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=300_h=200_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=300_h=200_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=300_h=200_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=300_h=200_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=300_h=200_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=300_h=200_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=300_h=200_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=360_h=270_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=360_h=270_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=360_h=270_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=360_h=270_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=360_h=270_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=360_h=270_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=360_h=270_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=360_h=270_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=360_h=270_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=360_h=270_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=360_h=270_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=360_h=270_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=360_h=270_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=360_h=270_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.jpeg", "w=360_h=270_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.jpeg", "w=360_h=270_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=100_h=150_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=100_h=150_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=100_h=150_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=100_h=150_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=100_h=150_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=100_h=150_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=100_h=150_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=100_h=150_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=100_h=150_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=100_h=150_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=100_h=150_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=100_h=150_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=100_h=150_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=100_h=150_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=100_h=150_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=100_h=150_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=240_h=240_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=240_h=240_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=240_h=240_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=240_h=240_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=240_h=240_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=240_h=240_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=240_h=240_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=240_h=240_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=240_h=240_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=240_h=240_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=240_h=240_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=240_h=240_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=240_h=240_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=240_h=240_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=240_h=240_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=240_h=240_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=250_h=250_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=250_h=250_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=250_h=250_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=250_h=250_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=250_h=250_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=250_h=250_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=250_h=250_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=250_h=250_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=250_h=250_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=250_h=250_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=250_h=250_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=250_h=250_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=250_h=250_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=250_h=250_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=250_h=250_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=250_h=250_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=300_h=200_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=300_h=200_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=300_h=200_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=300_h=200_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=300_h=200_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=300_h=200_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=300_h=200_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=300_h=200_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=300_h=200_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=300_h=200_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=300_h=200_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=300_h=200_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=300_h=200_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=300_h=200_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=300_h=200_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=300_h=200_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=360_h=270_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=360_h=270_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=360_h=270_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=360_h=270_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=360_h=270_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=360_h=270_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=360_h=270_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=360_h=270_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=360_h=270_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=360_h=270_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=360_h=270_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=360_h=270_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=360_h=270_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=360_h=270_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_180_240.png", "w=360_h=270_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_180_240.png", "w=360_h=270_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=100_h=150_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=100_h=150_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=100_h=150_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=100_h=150_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=100_h=150_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=100_h=150_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=100_h=150_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=100_h=150_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=100_h=150_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=100_h=150_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=100_h=150_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=100_h=150_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=100_h=150_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=100_h=150_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=100_h=150_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=100_h=150_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=240_h=240_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=240_h=240_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=240_h=240_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=240_h=240_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=240_h=240_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=240_h=240_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=240_h=240_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=240_h=240_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=240_h=240_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=240_h=240_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=240_h=240_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=240_h=240_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=240_h=240_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=240_h=240_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=240_h=240_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=240_h=240_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=250_h=250_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=250_h=250_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=250_h=250_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=250_h=250_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=250_h=250_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=250_h=250_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=250_h=250_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=250_h=250_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=250_h=250_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=250_h=250_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=250_h=250_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=250_h=250_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=250_h=250_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=250_h=250_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=250_h=250_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=250_h=250_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=300_h=200_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=300_h=200_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=300_h=200_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=300_h=200_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=300_h=200_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=300_h=200_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=300_h=200_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=300_h=200_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=300_h=200_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=300_h=200_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=300_h=200_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=300_h=200_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=300_h=200_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=300_h=200_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=300_h=200_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=300_h=200_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=360_h=270_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=360_h=270_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=360_h=270_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=360_h=270_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=360_h=270_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=360_h=270_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=360_h=270_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=360_h=270_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=360_h=270_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=360_h=270_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=360_h=270_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=360_h=270_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=360_h=270_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=360_h=270_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.jpeg", "w=360_h=270_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.jpeg", "w=360_h=270_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=100_h=150_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=100_h=150_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=100_h=150_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=100_h=150_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=100_h=150_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=100_h=150_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=100_h=150_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=100_h=150_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=100_h=150_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=100_h=150_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=100_h=150_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=100_h=150_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=100_h=150_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=100_h=150_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=100_h=150_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=100_h=150_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=240_h=240_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=240_h=240_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=240_h=240_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=240_h=240_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=240_h=240_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=240_h=240_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=240_h=240_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=240_h=240_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=240_h=240_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=240_h=240_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=240_h=240_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=240_h=240_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=240_h=240_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=240_h=240_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=240_h=240_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=240_h=240_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=250_h=250_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=250_h=250_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=250_h=250_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=250_h=250_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=250_h=250_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=250_h=250_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=250_h=250_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=250_h=250_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=250_h=250_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=250_h=250_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=250_h=250_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=250_h=250_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=250_h=250_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=250_h=250_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=250_h=250_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=250_h=250_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=300_h=200_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=300_h=200_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=300_h=200_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=300_h=200_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=300_h=200_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=300_h=200_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=300_h=200_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=300_h=200_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=300_h=200_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=300_h=200_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=300_h=200_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=300_h=200_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=300_h=200_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=300_h=200_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=300_h=200_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=300_h=200_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=360_h=270_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=360_h=270_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=360_h=270_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=360_h=270_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=360_h=270_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=360_h=270_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=360_h=270_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=360_h=270_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=360_h=270_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=360_h=270_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=360_h=270_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=360_h=270_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=360_h=270_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=360_h=270_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_180.png", "w=360_h=270_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_180.png", "w=360_h=270_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=100_h=150_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=100_h=150_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=100_h=150_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=100_h=150_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=100_h=150_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=100_h=150_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=100_h=150_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=100_h=150_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=100_h=150_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=100_h=150_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=100_h=150_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=100_h=150_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=100_h=150_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=100_h=150_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=100_h=150_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=100_h=150_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=240_h=240_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=240_h=240_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=240_h=240_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=240_h=240_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=240_h=240_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=240_h=240_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=240_h=240_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=240_h=240_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=240_h=240_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=240_h=240_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=240_h=240_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=240_h=240_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=240_h=240_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=240_h=240_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=240_h=240_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=240_h=240_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=250_h=250_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=250_h=250_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=250_h=250_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=250_h=250_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=250_h=250_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=250_h=250_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=250_h=250_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=250_h=250_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=250_h=250_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=250_h=250_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=250_h=250_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=250_h=250_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=250_h=250_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=250_h=250_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=250_h=250_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=250_h=250_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=300_h=200_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=300_h=200_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=300_h=200_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=300_h=200_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=300_h=200_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=300_h=200_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=300_h=200_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=300_h=200_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=300_h=200_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=300_h=200_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=300_h=200_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=300_h=200_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=300_h=200_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=300_h=200_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=300_h=200_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=300_h=200_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=360_h=270_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=360_h=270_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=360_h=270_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=360_h=270_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=360_h=270_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=360_h=270_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=360_h=270_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=360_h=270_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=360_h=270_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=360_h=270_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=360_h=270_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=360_h=270_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=360_h=270_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=360_h=270_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.jpeg", "w=360_h=270_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.jpeg", "w=360_h=270_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=100_h=150_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=100_h=150_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=100_h=150_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=100_h=150_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=100_h=150_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=100_h=150_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=100_h=150_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=100_h=150_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=100_h=150_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=100_h=150_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=100_h=150_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=100_h=150_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=100_h=150_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=100_h=150_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=100_h=150_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=100_h=150_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=240_h=240_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=240_h=240_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=240_h=240_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=240_h=240_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=240_h=240_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=240_h=240_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=240_h=240_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=240_h=240_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=240_h=240_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=240_h=240_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=240_h=240_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=240_h=240_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=240_h=240_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=240_h=240_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=240_h=240_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=240_h=240_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=250_h=250_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=250_h=250_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=250_h=250_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=250_h=250_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=250_h=250_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=250_h=250_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=250_h=250_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=250_h=250_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=250_h=250_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=250_h=250_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=250_h=250_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=250_h=250_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=250_h=250_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=250_h=250_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=250_h=250_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=250_h=250_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=300_h=200_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=300_h=200_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=300_h=200_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=300_h=200_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=300_h=200_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=300_h=200_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=300_h=200_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=300_h=200_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=300_h=200_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=300_h=200_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=300_h=200_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=300_h=200_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=300_h=200_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=300_h=200_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=300_h=200_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=300_h=200_m=scaleDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=360_h=270_m=fill_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=360_h=270_m=fill_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=360_h=270_m=fillDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=360_h=270_m=fillDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=360_h=270_m=fit_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=360_h=270_m=fit_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=360_h=270_m=fitDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=360_h=270_m=fitDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=360_h=270_m=pad_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=360_h=270_m=pad_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=360_h=270_m=padDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=360_h=270_m=padDown_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=360_h=270_m=scale_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=360_h=270_m=scale_b=ff00ffff_q=80.jpeg"),
					Entry("", "astronaut_240_240.png", "w=360_h=270_m=scaleDown_b=ff00ffff.png"),
					Entry("", "astronaut_240_240.png", "w=360_h=270_m=scaleDown_b=ff00ffff_q=80.jpeg"),
				)
			})
		})
	})
})

var workingDirectory = test.MustString(os.Getwd())
var fixturesDirectory = filepath.Join(workingDirectory, "test", "fixtures")
var renditionsDirectory = filepath.Join(fixturesDirectory, "renditions")
var sourcesDirectory = filepath.Join(fixturesDirectory, "sources")

var sourceBaseRegexp = regexp.MustCompile("^.*_([0-9]+)_([0-9]+).[^.]*$")
