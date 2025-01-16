package logger_test

import (
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/logger"
)

type Serializer struct {
	SerializeInvocations int
	SerializeInputs      []log.Fields
	SerializeOutputs     []error
}

func NewSerializer() *Serializer {
	return &Serializer{}
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
			lgr, err := logger.New(nil, log.DefaultLevelRanks(), log.DefaultLevel())
			Expect(err).To(MatchError("serializer is missing"))
			Expect(lgr).To(BeNil())
		})

		It("returns an error if the level ranks is missing", func() {
			lgr, err := logger.New(serializer, nil, log.DefaultLevel())
			Expect(err).To(MatchError("level ranks is missing"))
			Expect(lgr).To(BeNil())
		})

		It("returns an error if the level is not found", func() {
			lgr, err := logger.New(serializer, log.DefaultLevelRanks(), log.Level("unknown"))
			Expect(err).To(MatchError("level not found"))
			Expect(lgr).To(BeNil())
		})

		It("returns successfully", func() {
			Expect(logger.New(serializer, log.DefaultLevelRanks(), log.DefaultLevel())).ToNot(BeNil())
		})
	})

	Context("with new logger", func() {
		var lgr log.Logger

		BeforeEach(func() {
			var err error
			lgr, err = logger.New(serializer, log.DefaultLevelRanks(), log.DefaultLevel())
			Expect(err).ToNot(HaveOccurred())
			Expect(lgr).ToNot(BeNil())
		})

		Context("Log", func() {
			It("does not invoke serializer if the level is unknown", func() {
				lgr.Log(log.Level("unknown"), "Unknown Level Message")
			})

			It("does not invoke serializer if the level is not logging", func() {
				lgr.Log(log.DebugLevel, "Not Logging Message")
			})

			Context("with disabled standard error", func() {
				var newFile *os.File
				var oldFile *os.File

				BeforeEach(func() {
					var err error
					newFile, err = os.CreateTemp("", "")
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
					lgr.Log(log.WarnLevel, "Serializer Error Message")
					Expect(serializer.SerializeInputs).ToNot(BeEmpty())
				})
			})

			It("includes the expected fields in the expected format", func() {
				serializer.SerializeOutputs = []error{nil}
				lgr.Log(log.WarnLevel, "Expected Message")
				Expect(serializer.SerializeInputs).To(HaveLen(1))
				serializeInput := serializer.SerializeInputs[0]
				Expect(serializeInput).To(HaveKey("caller"))
				Expect(serializeInput).To(HaveKeyWithValue("level", log.WarnLevel))
				Expect(serializeInput).To(HaveKey("time"))
				Expect(serializeInput).To(HaveKeyWithValue("message", "Expected Message"))
				serializedTime, ok := serializeInput["time"].(string)
				Expect(ok).To(BeTrue())
				parsedTime, err := time.Parse(time.RFC3339Nano, serializedTime)
				Expect(err).ToNot(HaveOccurred())
				Expect(parsedTime).To(BeTemporally("~", time.Now(), time.Second))
				serializedCaller, ok := serializeInput["caller"].(*errors.Caller)
				Expect(ok).To(BeTrue())
				Expect(serializedCaller.Line).To(BeNumerically(">", 0))
				Expect(strings.HasSuffix(serializedCaller.File, "log/logger/logger_test.go")).To(BeTrue())
			})

			It("does not include the message is it is an empty string", func() {
				serializer.SerializeOutputs = []error{nil}
				lgr.Log(log.WarnLevel, "")
				Expect(serializer.SerializeInputs).To(HaveLen(1))
				Expect(serializer.SerializeInputs[0]).ToNot(HaveKey("message"))
			})
		})

		Context("with successful serialize and debug level", func() {
			BeforeEach(func() {
				serializer.SerializeOutputs = []error{nil}
				lgr = lgr.WithLevel(log.DebugLevel)
				Expect(lgr).ToNot(BeNil())
			})

			Context("Debug", func() {
				It("logs with the expected level and message", func() {
					lgr.Debug("Amazonian")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", log.DebugLevel))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "Amazonian"))
				})
			})

			Context("Info", func() {
				It("logs with the expected level and message", func() {
					lgr.Info("Bostonian")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", log.InfoLevel))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "Bostonian"))
				})
			})

			Context("Warn", func() {
				It("logs with the expected level and message", func() {
					lgr.Warn("Canadian")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", log.WarnLevel))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "Canadian"))
				})
			})

			Context("Error", func() {
				It("logs with the expected level and message", func() {
					lgr.Error("Dutch")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", log.ErrorLevel))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "Dutch"))
				})
			})

			Context("Debugf", func() {
				It("logs with the expected level and message", func() {
					lgr.Debugf("Amazonian %s", "Warrior")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", log.DebugLevel))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "Amazonian Warrior"))
				})
			})

			Context("Infof", func() {
				It("logs with the expected level and message", func() {
					lgr.Infof("Bostonian %s", "Cabbie")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", log.InfoLevel))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "Bostonian Cabbie"))
				})
			})

			Context("Warnf", func() {
				It("logs with the expected level and message", func() {
					lgr.Warnf("Canadian %s", "Skater")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", log.WarnLevel))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "Canadian Skater"))
				})
			})

			Context("Errorf", func() {
				It("logs with the expected level and message", func() {
					lgr.Errorf("Dutch %s", "Brothers")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", log.ErrorLevel))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "Dutch Brothers"))
				})
			})

			Context("WithError", func() {
				It("does not include the error field if the error is missing", func() {
					lgr.WithError(nil).Warn("European")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).ToNot(HaveKey("error"))
				})

				It("deletes the error field if the error is missing", func() {
					lgr.WithError(errors.New("euro error")).WithError(nil).Warn("European")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).ToNot(HaveKey("error"))
				})

				It("does include the error field if the error is not missing", func() {
					lgr.WithError(errors.New("euro error")).Warn("European")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKey("error"))
				})
			})

			Context("WithField", func() {
				It("does not include the field if the key is missing", func() {
					lgr.WithField("", "fish").Warn("Finnish")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).ToNot(HaveKey(""))
				})

				It("does not include the field if the value is missing", func() {
					lgr.WithField("sword", nil).Warn("Finnish")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).ToNot(HaveKey("sword"))
				})

				It("deletes the field if the value is missing", func() {
					lgr.WithField("sword", "fish").WithField("sword", nil).Warn("Finnish")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).ToNot(HaveKey("sword"))
				})

				It("does include the field if the key and value are not missing", func() {
					lgr.WithField("sword", "fish").Warn("Finnish")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("sword", "fish"))
				})
			})

			Context("WithFields", func() {
				It("does include the field if the key and value are not missing", func() {
					lgr.WithFields(log.Fields{"": "Nein", "nope": nil, "yep": "Ja"}).Warn("German")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).ToNot(HaveKey(""))
					Expect(serializer.SerializeInputs[0]).ToNot(HaveKey("nope"))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("yep", "Ja"))
				})

				It("deletes the field if the value is missing", func() {
					lgr.WithFields(log.Fields{"nope": "Nein"}).WithFields(log.Fields{"nope": nil}).Warn("German")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).ToNot(HaveKey("nope"))
				})
			})

			Context("WithLevelRank", func() {
				It("adds the specified level and rank", func() {
					level := log.Level("new")
					lgr = lgr.WithLevelRank(level, 90).WithLevel(level)
					Expect(lgr).ToNot(BeNil())
					lgr.Debug("Should Not Serialize")
					lgr.Log(level, "WithLevelRank Message")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", level))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "WithLevelRank Message"))
				})
			})

			Context("WithLevelRanks", func() {
				It("adds the specified level ranks", func() {
					level := log.Level("new")
					lgr = lgr.WithLevelRanks(log.LevelRanks{level: 30, log.Level("other"): 0}).WithLevel(level)
					Expect(lgr).ToNot(BeNil())
					lgr.Debug("Should Not Serialize")
					lgr.Log(level, "WithLevelRanks Message")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", level))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "WithLevelRanks Message"))
				})
			})

			Context("WithLevel", func() {
				It("uses the current level if the specified level is unknown", func() {
					lgr = lgr.WithLevel(log.Level("unknown"))
					Expect(lgr).ToNot(BeNil())
					Expect(lgr.Level()).To(Equal(log.DebugLevel))
					lgr.Debug("WithLevel Message")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", log.DebugLevel))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "WithLevel Message"))
				})

				It("adds the specified level", func() {
					lgr = lgr.WithLevel(log.InfoLevel)
					Expect(lgr).ToNot(BeNil())
					lgr.Debug("Should Not Serialize")
					lgr.Warn("WithLevel Message")
					Expect(serializer.SerializeInputs).To(HaveLen(1))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("level", log.WarnLevel))
					Expect(serializer.SerializeInputs[0]).To(HaveKeyWithValue("message", "WithLevel Message"))
				})
			})
		})

		Context("Level", func() {
			It("returns the current level", func() {
				Expect(lgr.Level()).To(Equal(log.WarnLevel))
			})

			It("returns the level after being set", func() {
				lgr = lgr.WithLevel(log.DebugLevel)
				Expect(lgr).ToNot(BeNil())
				Expect(lgr.Level()).To(Equal(log.DebugLevel))
			})

			It("returns the level after a new level is added and set", func() {
				level := log.Level("new")
				lgr = lgr.WithLevelRank(level, 55).WithLevel(level)
				Expect(lgr).ToNot(BeNil())
				Expect(lgr.Level()).To(Equal(level))
			})
		})
	})
})
