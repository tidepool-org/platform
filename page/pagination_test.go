package page_test

import (
	"net/http"
	"net/url"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	errorsTest "github.com/tidepool-org/platform/errors/test"
	logTest "github.com/tidepool-org/platform/log/test"
	"github.com/tidepool-org/platform/page"
	pageTest "github.com/tidepool-org/platform/page/test"
	structureParser "github.com/tidepool-org/platform/structure/parser"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
	"github.com/tidepool-org/platform/test"
	testHttp "github.com/tidepool-org/platform/test/http"
)

var _ = Describe("Pagination", func() {
	It("PaginationPageDefault is expected", func() {
		Expect(page.PaginationPageDefault).To(Equal(0))
	})

	It("PaginationPageMinimum is expected", func() {
		Expect(page.PaginationPageMinimum).To(Equal(0))
	})

	It("PaginationSizeDefault is expected", func() {
		Expect(page.PaginationSizeDefault).To(Equal(100))
	})

	It("PaginationSizeMaximum is expected", func() {
		Expect(page.PaginationSizeMaximum).To(Equal(1000))
	})

	It("PaginationSizeMinimum is expected", func() {
		Expect(page.PaginationSizeMinimum).To(Equal(1))
	})

	Context("Pagination", func() {
		Context("NewPagination", func() {
			It("returns successfully with default values", func() {
				datum := page.NewPagination()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Page).To(Equal(page.PaginationPageDefault))
				Expect(datum.Size).To(Equal(page.PaginationSizeDefault))
			})
		})

		Context("NewPaginationMinimum", func() {
			It("returns successfully with minimum values", func() {
				datum := page.NewPaginationMinimum()
				Expect(datum).ToNot(BeNil())
				Expect(datum.Page).To(Equal(page.PaginationPageMinimum))
				Expect(datum.Size).To(Equal(page.PaginationSizeMinimum))
			})
		})

		Context("Parse", func() {
			DescribeTable("parses the datum",
				func(mutator func(object map[string]any, expectedDatum *page.Pagination), expectedErrors ...error) {
					expectedDatum := pageTest.RandomPagination()
					object := pageTest.NewObjectFromPagination(expectedDatum, test.ObjectFormatJSON)
					mutator(object, expectedDatum)
					datum := &page.Pagination{}
					errorsTest.ExpectEqual(structureParser.NewObject(logTest.NewLogger(), &object).Parse(datum), expectedErrors...)
					Expect(datum).To(Equal(expectedDatum))
				},
				Entry("succeeds",
					func(object map[string]any, expectedDatum *page.Pagination) {},
				),
				Entry("page invalid type",
					func(object map[string]any, expectedDatum *page.Pagination) {
						object["page"] = true
						expectedDatum.Page = 0
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/page"),
				),
				Entry("page valid",
					func(object map[string]any, expectedDatum *page.Pagination) {
						valid := pageTest.RandomPage()
						object["page"] = valid
						expectedDatum.Page = valid
					},
				),
				Entry("size invalid type",
					func(object map[string]any, expectedDatum *page.Pagination) {
						object["size"] = true
						expectedDatum.Size = 0
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/size"),
				),
				Entry("size valid",
					func(object map[string]any, expectedDatum *page.Pagination) {
						valid := pageTest.RandomSize()
						object["size"] = valid
						expectedDatum.Size = valid
					},
				),
				Entry("multiple errors",
					func(object map[string]any, expectedDatum *page.Pagination) {
						object["page"] = true
						object["size"] = true
						expectedDatum.Page = 0
						expectedDatum.Size = 0
					},
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/page"),
					errorsTest.WithPointerSource(structureParser.ErrorTypeNotInt(true), "/size"),
				),
			)
		})

		Context("Validate", func() {
			DescribeTable("validates the datum",
				func(mutator func(datum *page.Pagination), expectedErrors ...error) {
					datum := pageTest.RandomPagination()
					mutator(datum)
					errorsTest.ExpectEqual(structureValidator.New(logTest.NewLogger()).Validate(datum), expectedErrors...)
				},
				Entry("succeeds",
					func(datum *page.Pagination) {},
				),
				Entry("page out of range (lower)",
					func(datum *page.Pagination) { datum.Page = page.PaginationPageMinimum - 1 },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(page.PaginationPageMinimum-1, page.PaginationPageMinimum), "/page"),
				),
				Entry("page valid",
					func(datum *page.Pagination) { datum.Page = pageTest.RandomPage() },
				),
				Entry("size out of range (lower)",
					func(datum *page.Pagination) { datum.Size = page.PaginationSizeMinimum - 1 },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(page.PaginationSizeMinimum-1, page.PaginationSizeMinimum, page.PaginationSizeMaximum), "/size"),
				),
				Entry("size out of range (upper)",
					func(datum *page.Pagination) { datum.Size = page.PaginationSizeMaximum + 1 },
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(page.PaginationSizeMaximum+1, page.PaginationSizeMinimum, page.PaginationSizeMaximum), "/size"),
				),
				Entry("size valid",
					func(datum *page.Pagination) { datum.Size = pageTest.RandomSize() },
				),
				Entry("multiple errors",
					func(datum *page.Pagination) {
						datum.Page = page.PaginationPageMinimum - 1
						datum.Size = page.PaginationSizeMinimum - 1
					},
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotGreaterThanOrEqualTo(page.PaginationPageMinimum-1, page.PaginationPageMinimum), "/page"),
					errorsTest.WithPointerSource(structureValidator.ErrorValueNotInRange(page.PaginationSizeMinimum-1, page.PaginationSizeMinimum, page.PaginationSizeMaximum), "/size"),
				),
			)
		})

		Context("with a new pagination", func() {
			var datum *page.Pagination

			BeforeEach(func() {
				datum = pageTest.RandomPagination()
			})

			Context("MutateRequest", func() {
				It("returns an error if the request is missing", func() {
					Expect(datum.MutateRequest(nil)).To(MatchError("request is missing"))
				})

				Context("with a request", func() {
					var request *http.Request

					BeforeEach(func() {
						request = testHttp.NewRequest()
					})

					It("does not add default page and size to the request as query parameters", func() {
						datum = page.NewPagination()
						Expect(datum.MutateRequest(request)).To(Succeed())
						Expect(request.URL.Query()).To(BeEmpty())
					})

					It("adds custom page and size to the request as query parameters", func() {
						Expect(datum.MutateRequest(request)).To(Succeed())
						Expect(request.URL.Query()).To(Equal(url.Values{"page": []string{strconv.Itoa(datum.Page)}, "size": []string{strconv.Itoa(datum.Size)}}))
					})
				})
			})
		})

		Context("Paginate", func() {
			It("returns error if paginator is missing", func() {
				err := page.Paginate(nil)
				Expect(err).To(MatchError("paginator is missing"))
			})

			It("returns error if paginator returns error", func() {
				err := errorsTest.RandomError()
				paginator := func(p page.Pagination) (bool, error) { return false, err }
				Expect(page.Paginate(paginator)).To(Equal(err))
			})

			It("calls paginator with default size and increments page until done", func() {
				expectedPages := make([]int, test.RandomIntFromRange(1, 10))
				for index := range expectedPages {
					expectedPages[index] = index
				}

				var actualPages []int
				paginator := func(p page.Pagination) (bool, error) {
					Expect(p.Size).To(Equal(page.PaginationSizeDefault))
					actualPages = append(actualPages, p.Page)
					return len(actualPages) >= len(expectedPages), nil
				}

				Expect(page.Paginate(paginator)).To(Succeed())
				Expect(actualPages).To(Equal(expectedPages))
			})
		})

		Context("PaginateWithSize", func() {
			It("returns error if size is less than minimum", func() {
				err := page.PaginateWithSize(page.PaginationSizeMinimum-1, func(p page.Pagination) (bool, error) { return false, nil })
				Expect(err).To(MatchError("size is less than minimum"))
			})

			It("returns error if paginator is missing", func() {
				err := page.PaginateWithSize(pageTest.RandomSize(), nil)
				Expect(err).To(MatchError("paginator is missing"))
			})

			It("returns error if paginator returns error", func() {
				err := errorsTest.RandomError()
				paginator := func(p page.Pagination) (bool, error) { return false, err }
				Expect(page.PaginateWithSize(pageTest.RandomSize(), paginator)).To(Equal(err))
			})

			It("calls paginator with given size and increments page until done", func() {
				size := pageTest.RandomSize()
				expectedPages := make([]int, test.RandomIntFromRange(1, 10))
				for index := range expectedPages {
					expectedPages[index] = index
				}

				var actualPages []int
				paginator := func(p page.Pagination) (bool, error) {
					Expect(p.Size).To(Equal(size))
					actualPages = append(actualPages, p.Page)
					return len(actualPages) >= len(expectedPages), nil
				}

				Expect(page.PaginateWithSize(size, paginator)).To(Succeed())
				Expect(actualPages).To(Equal(expectedPages))
			})
		})

		Context("Collect", func() {
			It("returns error if pager is missing", func() {
				result, err := page.Collect[string, []string](nil)
				Expect(err).To(MatchError("pager is missing"))
				Expect(result).To(BeNil())
			})

			It("returns error if pager returns error", func() {
				err := errorsTest.RandomError()
				pager := func(p page.Pagination) ([]string, error) { return nil, err }
				result, err := page.Collect(pager)
				Expect(err).To(Equal(err))
				Expect(result).To(BeNil())
			})

			It("calls pager with default size and returns nil elements", func() {
				pager := func(p page.Pagination) ([]string, error) {
					Expect(p.Size).To(Equal(page.PaginationSizeDefault))
					return nil, nil
				}

				actualCollection, err := page.Collect(pager)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualCollection).To(BeNil())
			})

			It("calls pager with default size and returns empty elements", func() {
				pager := func(p page.Pagination) ([]string, error) {
					Expect(p.Size).To(Equal(page.PaginationSizeDefault))
					return []string{}, nil
				}

				actualCollection, err := page.Collect(pager)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualCollection).To(Equal([]string{}))
			})

			It("calls pager with default size and increments page until done", func() {
				expectedCollection := make([]string, test.RandomIntFromRange(0, page.PaginationSizeDefault*test.RandomIntFromRange(1, 10)))
				for index := range expectedCollection {
					expectedCollection[index] = strconv.Itoa(test.RandomInt())
				}
				expectedPages := make([]int, (len(expectedCollection)+page.PaginationSizeDefault)/page.PaginationSizeDefault)
				for index := range expectedPages {
					expectedPages[index] = index
				}

				var actualPages []int
				pager := func(p page.Pagination) ([]string, error) {
					Expect(p.Size).To(Equal(page.PaginationSizeDefault))
					actualPages = append(actualPages, p.Page)
					startIndex := p.Page * p.Size
					endIndex := min(startIndex+p.Size, len(expectedCollection))
					return expectedCollection[startIndex:endIndex], nil
				}

				actualCollection, err := page.Collect(pager)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualCollection).To(Equal(expectedCollection))
				Expect(actualPages).To(Equal(expectedPages))
			})
		})

		Context("CollectWithSize", func() {
			It("returns error if size is less than minimum", func() {
				result, err := page.CollectWithSize(page.PaginationSizeMinimum-1, func(p page.Pagination) ([]string, error) { return nil, nil })
				Expect(err).To(MatchError("size is less than minimum"))
				Expect(result).To(BeNil())
			})

			It("returns error if pager is missing", func() {
				result, err := page.CollectWithSize[string, []string](pageTest.RandomSize(), nil)
				Expect(err).To(MatchError("pager is missing"))
				Expect(result).To(BeNil())
			})

			It("returns error if pager returns error", func() {
				err := errorsTest.RandomError()
				pager := func(p page.Pagination) ([]string, error) { return nil, err }
				result, err := page.CollectWithSize(pageTest.RandomSize(), pager)
				Expect(err).To(Equal(err))
				Expect(result).To(BeNil())
			})

			It("calls pager with given size and returns nil elements", func() {
				size := pageTest.RandomSize()

				pager := func(p page.Pagination) ([]string, error) {
					Expect(p.Size).To(Equal(size))
					return nil, nil
				}

				actualCollection, err := page.CollectWithSize(size, pager)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualCollection).To(BeNil())
			})

			It("calls pager with given size and returns empty elements", func() {
				size := pageTest.RandomSize()

				pager := func(p page.Pagination) ([]string, error) {
					Expect(p.Size).To(Equal(size))
					return []string{}, nil
				}

				actualCollection, err := page.CollectWithSize(size, pager)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualCollection).To(Equal([]string{}))
			})

			It("calls pager with given size and increments page until done", func() {
				size := pageTest.RandomSize()
				expectedCollection := make([]string, test.RandomIntFromRange(0, size*test.RandomIntFromRange(1, 10)))
				for index := range expectedCollection {
					expectedCollection[index] = strconv.Itoa(test.RandomInt())
				}
				expectedPages := make([]int, (len(expectedCollection)+size)/size)
				for index := range expectedPages {
					expectedPages[index] = index
				}

				var actualPages []int
				pager := func(p page.Pagination) ([]string, error) {
					Expect(p.Size).To(Equal(size))
					actualPages = append(actualPages, p.Page)
					startIndex := p.Page * p.Size
					endIndex := min(startIndex+p.Size, len(expectedCollection))
					return expectedCollection[startIndex:endIndex], nil
				}

				actualCollection, err := page.CollectWithSize(size, pager)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualCollection).To(Equal(expectedCollection))
				Expect(actualPages).To(Equal(expectedPages))
			})
		})

		Context("Process", func() {
			It("returns error if pager is missing", func() {
				processor := func(element int) (string, error) { return strconv.Itoa(element), nil }
				result, err := page.Process[int, string, []int](nil, processor)
				Expect(err).To(MatchError("pager is missing"))
				Expect(result).To(BeNil())
			})

			It("returns error if processor is missing", func() {
				pager := func(p page.Pagination) ([]int, error) { return nil, nil }
				result, err := page.Process[int, string](pager, nil)
				Expect(err).To(MatchError("processor is missing"))
				Expect(result).To(BeNil())
			})

			It("returns error if pager returns error", func() {
				err := errorsTest.RandomError()
				pager := func(p page.Pagination) ([]int, error) { return nil, err }
				processor := func(element int) (string, error) { return strconv.Itoa(element), nil }
				result, err := page.Process(pager, processor)
				Expect(err).To(Equal(err))
				Expect(result).To(BeNil())
			})

			It("returns error if processor returns error", func() {
				err := errorsTest.RandomError()
				pager := func(p page.Pagination) ([]int, error) { return []int{1, 2, 3}, nil }
				processor := func(element int) (string, error) { return "", err }
				result, err := page.Process(pager, processor)
				Expect(err).To(Equal(err))
				Expect(result).To(BeEmpty())
			})

			It("calls pager with default size, nil elements returned", func() {
				pager := func(p page.Pagination) ([]int, error) {
					Expect(p.Size).To(Equal(page.PaginationSizeDefault))
					return nil, nil
				}
				processor := func(element int) (string, error) {
					return strconv.Itoa(element), nil
				}

				actualCollection, err := page.Process(pager, processor)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualCollection).To(BeNil())
			})

			It("calls pager with default size, empty elements returned", func() {
				pager := func(p page.Pagination) ([]int, error) {
					Expect(p.Size).To(Equal(page.PaginationSizeDefault))
					return []int{}, nil
				}
				processor := func(element int) (string, error) {
					return strconv.Itoa(element), nil
				}

				actualCollection, err := page.Process(pager, processor)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualCollection).To(Equal([]string{}))
			})

			It("calls pager with default size, processes each element, and increments page until done", func() {
				pagedCollection := make([]int, test.RandomIntFromRange(0, page.PaginationSizeDefault*test.RandomIntFromRange(1, 10)))
				for index := range pagedCollection {
					pagedCollection[index] = test.RandomInt()
				}
				expectedCollection := make([]string, len(pagedCollection))
				for index := range expectedCollection {
					expectedCollection[index] = strconv.Itoa(pagedCollection[index])
				}
				expectedPages := make([]int, (len(pagedCollection)+page.PaginationSizeDefault)/page.PaginationSizeDefault)
				for index := range expectedPages {
					expectedPages[index] = index
				}

				var actualPages []int
				pager := func(p page.Pagination) ([]int, error) {
					Expect(p.Size).To(Equal(page.PaginationSizeDefault))
					actualPages = append(actualPages, p.Page)
					startIndex := p.Page * p.Size
					endIndex := min(startIndex+p.Size, len(pagedCollection))
					return pagedCollection[startIndex:endIndex], nil
				}
				processor := func(element int) (string, error) {
					return strconv.Itoa(element), nil
				}

				actualCollection, err := page.Process(pager, processor)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualCollection).To(Equal(expectedCollection))
				Expect(actualPages).To(Equal(expectedPages))
			})
		})

		Context("ProcessWithSize", func() {
			It("returns error if size is less than minimum", func() {
				pager := func(p page.Pagination) ([]int, error) { return nil, nil }
				processor := func(element int) (string, error) { return strconv.Itoa(element), nil }
				result, err := page.ProcessWithSize(page.PaginationSizeMinimum-1, pager, processor)
				Expect(err).To(MatchError("size is less than minimum"))
				Expect(result).To(BeNil())
			})

			It("returns error if pager is missing", func() {
				processor := func(element int) (string, error) { return strconv.Itoa(element), nil }
				result, err := page.ProcessWithSize[int, string, []int](pageTest.RandomSize(), nil, processor)
				Expect(err).To(MatchError("pager is missing"))
				Expect(result).To(BeNil())
			})

			It("returns error if processor is missing", func() {
				pager := func(p page.Pagination) ([]int, error) { return nil, nil }
				result, err := page.ProcessWithSize[int, string](pageTest.RandomSize(), pager, nil)
				Expect(err).To(MatchError("processor is missing"))
				Expect(result).To(BeNil())
			})

			It("returns error if pager returns error", func() {
				err := errorsTest.RandomError()
				pager := func(p page.Pagination) ([]int, error) { return nil, err }
				processor := func(element int) (string, error) { return strconv.Itoa(element), nil }
				result, err := page.ProcessWithSize(pageTest.RandomSize(), pager, processor)
				Expect(err).To(Equal(err))
				Expect(result).To(BeNil())
			})

			It("returns error if processor returns error", func() {
				err := errorsTest.RandomError()
				pager := func(p page.Pagination) ([]int, error) { return []int{1, 2, 3}, nil }
				processor := func(element int) (string, error) { return "", err }
				result, err := page.ProcessWithSize(pageTest.RandomSize(), pager, processor)
				Expect(err).To(Equal(err))
				Expect(result).To(BeEmpty())
			})

			It("calls pager with given size, nil elements returned", func() {
				size := pageTest.RandomSize()

				pager := func(p page.Pagination) ([]int, error) {
					Expect(p.Size).To(Equal(size))
					return nil, nil
				}
				processor := func(element int) (string, error) {
					return strconv.Itoa(element), nil
				}

				actualCollection, err := page.ProcessWithSize(size, pager, processor)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualCollection).To(BeNil())
			})

			It("calls pager with given size, empty elements returned", func() {
				size := pageTest.RandomSize()

				pager := func(p page.Pagination) ([]int, error) {
					Expect(p.Size).To(Equal(size))
					return []int{}, nil
				}
				processor := func(element int) (string, error) {
					return strconv.Itoa(element), nil
				}

				actualCollection, err := page.ProcessWithSize(size, pager, processor)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualCollection).To(Equal([]string{}))
			})

			It("calls pager with given size, processes each element, and increments page until done", func() {
				size := pageTest.RandomSize()
				pagedCollection := make([]int, test.RandomIntFromRange(0, size*test.RandomIntFromRange(1, 10)))
				for index := range pagedCollection {
					pagedCollection[index] = test.RandomInt()
				}
				expectedCollection := make([]string, len(pagedCollection))
				for index := range expectedCollection {
					expectedCollection[index] = strconv.Itoa(pagedCollection[index])
				}
				expectedPages := make([]int, (len(pagedCollection)+size)/size)
				for index := range expectedPages {
					expectedPages[index] = index
				}

				var actualPages []int
				pager := func(p page.Pagination) ([]int, error) {
					Expect(p.Size).To(Equal(size))
					actualPages = append(actualPages, p.Page)
					startIndex := p.Page * p.Size
					endIndex := min(startIndex+p.Size, len(pagedCollection))
					return pagedCollection[startIndex:endIndex], nil
				}
				processor := func(element int) (string, error) {
					return strconv.Itoa(element), nil
				}

				actualCollection, err := page.ProcessWithSize(size, pager, processor)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualCollection).To(Equal(expectedCollection))
				Expect(actualPages).To(Equal(expectedPages))
			})
		})
	})
})
