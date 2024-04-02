package buffer

import "testing"

func TestSetupProducer(t *testing.T) {
	producer, err := SetupProducer()
	if err != nil {
		t.Errorf("failed to setup producer: %v", err)
	}
	if producer == nil {
		t.Errorf("failed to setup producer")
	}

}
