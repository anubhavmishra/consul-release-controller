package models

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/nicholasjackson/consul-release-controller/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSetsUpPluginsAndState(t *testing.T) {
	// test vanilla
	d := &Release{}
	data := bytes.NewBuffer(testutils.GetTestData(t, "valid_kubernetes_release.json"))
	err := d.FromJsonBody(ioutil.NopCloser(data))
	assert.NoError(t, err)
}

func TestToJsonSerializesState(t *testing.T) {
	d := &Release{}
	data := bytes.NewBuffer(testutils.GetTestData(t, "valid_kubernetes_release.json"))
	err := d.FromJsonBody(ioutil.NopCloser(data))
	assert.NoError(t, err)

	releaseJson := d.ToJson()
	require.Contains(t, string(releaseJson), `"name":"api"`)
}
