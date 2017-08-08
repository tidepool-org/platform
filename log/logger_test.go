package log_test

import (
	"io/ioutil"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"
	"time"

	"github.com/tidepool-org/platform/id"
	"github.com/tidepool-org/platform/log"
)

type Serializer struct {
	ID                   string
	SerializeInvocations int
	SerializeInputs      []log.Fields
	SerializeOutputs     []error
}

func NewSerializer() *Serializer {
	return &Serializer{
		ID: id.New(),
	}
}

func (s *Serializer) Serialize(fields log.Fields) error {
	s.SerializeInvocations++

	s.SerializeInputs = append(s.SerializeInputs, fields)

	if len(s.SerializeOutputs) == 0 {
		panic("Unexpected invocation of Serialize on Serializer")
	}

	output := s.SerializeOutputs[0]
	s.SerializeOutputs = s.SerializeOutputs[1:]
	return output
}

func (s *Serializer) UnusedOutputsCount() int {
	return len(s.SerializeOutputs)
}

var _ = Describe("Logger", func() {
	var serializer *Serializer

	BeforeEach(func() {
		serializer = NewSerializer()
		Expect(serializer).ToNot(BeNil())
	})

	AfterEach(func() {
		Expect(serializer.UnusedOutputsCount()).To(Equal(0))
	})

	Context("NewLogger", func() {
		It("returns an error if the serializer is missing", func() {
			logger, err := log.NewLogger(nil, log.DefaultLevels(), log.DefaultLevel())
			Expect(err).To(MatchError("log: serializer is missing"))
			Expect(logger).To(BeNil())
		})

		It("returns an error if the levels is missing", func() {
			logger, err := log.NewLogger(serializer, nil, log.DefaultLevel())
			Expect(err).To(MatchError("log: levels is missing"))
			Expect(logger).To(BeNil())
		})

		It("returns an error if the level is not found", func() {
			logger, err := log.NewLogger(serializer, log.DefaultLevels(), log.Level("unknown"))
			Expect(err).To(MatchError("log: level not found"))
			Expect(logger).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(log.NewLogger(serializer, log.DefaultLevels(), log.DefaultLevel())).ToNot(BeNil())
		})
	})

	Context("with new logger", func() {
		var logger log.Logger

		BeforeEach(func() {
			var err error
			logger, err = log.NewLogger(serializer, log.DefaultLevels(), log.DefaultLevel())
			Expect(err).ToNot(HaveOccurred())
			Expect(logger).ToNot(BeNil())
		})

		Context("Log", func() {
			It("does not invoke serializer if the level is unknown", func() {
				logger.Log(log.Level("unknown"), "Unknown Level Message")
			})

			It("does not invoke serializer if the level is not logging", func() {
				logger.Log(log.DebugLevel, "Not Logging Message")
			})

			Context("with disabled standard error", func() {
				var newFile *os.File
				var oldFile *os.File

				BeforeEach(func() {
					var err error
					newFile, err = ioutil.TempFile("", "")
					Expect(err).ToNot(HaveOccurred())
					Expect(newFile).ToNot(BeNil())
					oldFile = os.Stderr
					os.Stderr = newFile
				})

				AfterEach(func() {
					os.Stderr = oldFile
					os.Remove(newFile.Name())
				})

				It("fails silently if the serializer returns an error", func() {
					serializer.SerializeOutputs = []error{errors.New("test error")}
					logger.Log(log.WarnLevel, "Serializer Error Message")
					Expect(serializer.SerializeInputs).ToNot(BeEmpty())
				})
			})

			It("includes the expected fields in the expected format", func() {
				serializer.SerializeOutputs = []error{nil}
				logger.Log(log.WarnLevel, "Expected Message")
				Expect(serializer.SerializeInputs).To(HaveLen(1))
				serializeInput := serializer.SerializeInputs[0]
				Expect(serializeInput).To(HaveKey("time"))
				Expect(serializeInput).To(HaveKey("line"))
				Expect(serializeInput).To(HaveKey("file"))
				Expect(serializeInput).To(HaveKeyWithValue("level", log.WarnLevel))
				Expect(serializeInput).To(HaveKeyWithValue("message", "Expected Message"))
				serializedTime, ok := serializeInput["time"].(string)
				Expect(ok).To(BeTrue())
				parsedTime, err := time.Parse("2006-01-02T15:04:05.999Z07:00", serializedTime)
				Expect(err).ToNot(HaveOccurred())
				Expect(parsedTime).To(BeTemporally("~", time.Now(), time.Second))
				serializedLine := serializeInput["line"]
				Expect(serializedLine).To(BeNumerically(">", 0))
				serializedFile, ok := serializeInput["file"].(string)
				Expect(ok).To(BeTrue())
				Expect(strings.HasSuffix(serializedFile, "log/logger_test.go")).To(BeTrue())
			})

			It("does not include the message is it is an empty string", func() {
				serializer.SerializeOutputs = []error{nil}
				logger.Log(log.WarnLevel, "")
				Expect(serializer.SerializeInputs).To(HaveLen(1))
				Expect(serializer.SerializeInputs[0]).ToNot(HaveKey("message"))
			})
		})

		Context("with successful serialize and debug level", func() {
			BeforeEach(func() {
				serializer.SerializeOutputs = []error{nil}
				logger.SetLevel(log.DebugLevel)
			})

			Context("Debug", func() {
				It("logs with the expected level and message", func() {
					logger.Debug("Amazonian")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", log.DebugLevel))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "Amazonian"))
				})
			})

			Context("Info", func() {
				It("logs with the expected level and message", func() {
					logger.Info("Bostonian")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", log.InfoLevel))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "Bostonian"))
				})
			})

			Context("Warn", func() {
				It("logs with the expected level and message", func() {
					logger.Warn("Canadian")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", log.WarnLevel))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "Canadian"))
				})
			})

			Context("Error", func() {
				It("logs with the expected level and message", func() {
					logger.Error("Dutch")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", log.ErrorLevel))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "Dutch"))
				})
			})

			Context("Debugf", func() {
				It("logs with the expected level and message", func() {
					logger.Debugf("Amazonian %s", "Warrior")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", log.DebugLevel))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "Amazonian Warrior"))
				})
			})

			Context("Infof", func() {
				It("logs with the expected level and message", func() {
					logger.Infof("Bostonian %s", "Cabbie")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", log.InfoLevel))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "Bostonian Cabbie"))
				})
			})

			Context("Warnf", func() {
				It("logs with the expected level and message", func() {
					logger.Warnf("Canadian %s", "Skater")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", log.WarnLevel))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "Canadian Skater"))
				})
			})

			Context("Errorf", func() {
				It("logs with the expected level and message", func() {
					logger.Errorf("Dutch %s", "Brothers")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", log.ErrorLevel))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "Dutch Brothers"))
				})
			})

			Context("WithError", func() {
				It("does not include the error field if the error is nil", func() {
					logger.WithError(nil).Warn("European")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).ToNot(HaveKey("error"))
				})

				It("does include the error field if the error is not nil", func() {
					logger.WithError(errors.New("euro error")).Warn("European")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("error", "euro error"))
				})
			})

			Context("WithField", func() {
				It("does not include the field if the key is missing", func() {
					logger.WithField("", "fish").Warn("Finnish")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).ToNot(HaveKey(""))
				})

				It("does not include the field if the value is missing", func() {
					logger.WithField("sword", nil).Warn("Finnish")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).ToNot(HaveKey("sword"))
				})

				It("does include the field if the key and value are not missing", func() {
					logger.WithField("sword", "fish").Warn("Finnish")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("sword", "fish"))
				})
			})

			Context("WithFields", func() {
				It("does include the field if the key and value are not missing", func() {
					logger.WithFields(log.Fields{"": "Nein", "nope": nil, "yep": "Ja"}).Warn("German")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).ToNot(HaveKey(""))
					Expect(serializer.SerializeInputs[0]).ToNot(HaveKey("nope"))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("yep", "Ja"))
				})
			})

			Context("WithLevel", func() {
				It("adds the specified level", func() {
					level := log.Level("new")
					Expect(logger.SetLevel(level)).To(MatchError("log: level not found"))
					logger = logger.WithLevel(level, 90)
					Expect(logger.SetLevel(level)).To(Succeed())
					logger.Debug("Should Not Serialize")
					logger.Log(level, "WithLevel Message")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", level))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "WithLevel Message"))
				})
			})

			Context("WithLevels", func() {
				It("adds the specified level", func() {
					level := log.Level("new")
					Expect(logger.SetLevel(level)).To(MatchError("log: level not found"))
					logger = logger.WithLevels(log.Levels{level: 30, log.Level("other"): 0})
					Expect(logger.SetLevel(level)).To(Succeed())
					logger.Debug("Should Not Serialize")
					logger.Log(level, "WithLevels Message")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", level))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "WithLevels Message"))
				})
			})
		})

		Context("Level", func() {
			It("returns the current level", func() {
				Expect(logger.Level()).To(Equal(log.WarnLevel))
			})

			It("returns the level after being set", func() {
				Expect(logger.SetLevel(log.DebugLevel)).To(Succeed())
				Expect(logger.Level()).To(Equal(log.DebugLevel))
			})

			It("returns the level after a new level is added and set", func() {
				level := log.Level("new")
				logger = logger.WithLevel(level, 55)
				Expect(logger.SetLevel(level)).To(Succeed())
				Expect(logger.Level()).To(Equal(level))
			})
		})

		Context("SetLevel", func() {
			It("returns an error if the level is not found", func() {
				Expect(logger.SetLevel(log.Level("not found"))).To(MatchError("log: level not found"))
			})

			It("sets a new level", func() {
				Expect(logger.SetLevel(log.InfoLevel)).To(Succeed())
				Expect(logger.Level()).To(Equal(log.InfoLevel))
			})

			It("sets a new level that was just added", func() {
				level := log.Level("new")
				logger = logger.WithLevel(level, 77)
				Expect(logger.SetLevel(level)).To(Succeed())
				Expect(logger.Level()).To(Equal(level))
			})
		})
	})
})
