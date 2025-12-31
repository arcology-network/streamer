package jetlib

import (
	"testing"
)

func TestLoadCfg(t *testing.T) {
	cfg, err := LoadConfig("./jet.yaml")
	if err != nil {
		t.Error(err)
	}

	if v, ok := cfg.TopicsD["pendingblock"]; !ok || v.SubType != Broadcast {
		t.Error("not found pendingblock")
	}

	if v, ok := cfg.TopicsD["message"]; !ok || v.SubType != LoadBalance {
		t.Error("not found message")
	}

	if len(cfg.TopicsU) != 2 {
		t.Error("not found topic for upload")
	}

}
