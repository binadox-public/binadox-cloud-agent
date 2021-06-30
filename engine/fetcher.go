package engine

import (
	"encoding/json"
	"github.com/peterbourgon/diskv"
)

type InstanceInfo struct {
	Version string `json:"version"`
	Instance InstanceID `json:"instance"`
	Stat MachineStats `json:"stat"`
}

type FetcherContext struct {
	diskv *diskv.Diskv
}

func InitFetcher(storagePath string) FetcherContext {
	flatTransform := func(s string) []string { return []string{} }

	// Initialize a new diskv store, rooted at "my-data-dir", with a 1MB cache.
	var d* diskv.Diskv;
	d = diskv.New(diskv.Options{
		BasePath:     storagePath,
		Transform:    flatTransform,
		CacheSizeMax: 1024 * 1024,
	})
	return FetcherContext{diskv: d}
}

const INSTANCE_ID_KEY = "instanceID"
const CURRENT_APP = "currentApp"

func GetLatestApplication(ctx *FetcherContext) (string, error) {
	if ctx.diskv.Has(CURRENT_APP) {
		data, err := ctx.diskv.Read(CURRENT_APP)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return "", nil
}

func SetLatestApplication(ctx *FetcherContext, path string) error {
	b := []byte(path)
	return ctx.diskv.Write(CURRENT_APP, b)
}

func Fetch(ctx *FetcherContext) (*InstanceInfo, error) {
	var err error
	var data []byte
	var result InstanceInfo
	if ctx.diskv.Has(INSTANCE_ID_KEY) {
		data, err = ctx.diskv.Read(INSTANCE_ID_KEY)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, &result.Instance)
		if err != nil {
			return nil, err
		}
	} else {
		instId := GetInstanceID()
		if instId != nil {
			result.Instance = *instId
			data, err = json.MarshalIndent(result.Instance, "", " ")
			if err == nil {
				ctx.diskv.Write(INSTANCE_ID_KEY, data)
			}
		}
	}
	var stats *MachineStats
	stats, err = GetMachineStats()
	if err != nil {
		return nil, err
	}
	result.Stat = *stats
	return &result, nil
}

