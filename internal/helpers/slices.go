package helpers

import (
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rs/zerolog/log"
)

func SortByKeyTimeStamp(contents []types.Object, prefix string) {
	sort.Slice(contents, func(i, j int) bool {
		iKey := *contents[i].Key
		jKey := *contents[j].Key

		iTimeString := strings.TrimPrefix(iKey, prefix)
		jTimeString := strings.TrimPrefix(jKey, prefix)

		iTimeString = strings.TrimSuffix(iTimeString, ".gzip")
		jTimeString = strings.TrimSuffix(jTimeString, ".gzip")

		iTimeString = strings.TrimSuffix(iTimeString, ".archive")
		jTimeString = strings.TrimSuffix(jTimeString, ".archive")

		iTime, err := time.Parse(TimeFormat, iTimeString)
		if err != nil {
			log.Panic().Err(err).Msgf("Failed to parse time from %s", iKey)
			return false
		}

		jTime, err := time.Parse(TimeFormat, jTimeString)
		if err != nil {
			log.Panic().Err(err).Msgf("Failed to parse time from %s", jKey)
			return false
		}

		return iTime.Before(jTime)
	})
}
