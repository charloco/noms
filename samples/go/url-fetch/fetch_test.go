// Copyright 2016 Attic Labs, Inc. All rights reserved.
// Licensed under the Apache License, version 2.0:
// http://www.apache.org/licenses/LICENSE-2.0

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"gopkg.in/attic-labs/noms.v7/go/datas"
	"gopkg.in/attic-labs/noms.v7/go/spec"
	"gopkg.in/attic-labs/noms.v7/go/types"
	"gopkg.in/attic-labs/noms.v7/go/util/clienttest"
	"github.com/attic-labs/testify/suite"
)

func TestFetch(t *testing.T) {
	suite.Run(t, &testSuite{})
}

type testSuite struct {
	clienttest.ClientTestSuite
}

func (s *testSuite) TestImportFromStdin() {
	assert := s.Assert()

	oldStdin := os.Stdin
	newStdin, blobOut, err := os.Pipe()
	assert.NoError(err)

	os.Stdin = newStdin
	defer func() {
		os.Stdin = oldStdin
	}()

	go func() {
		blobOut.Write([]byte("abcdef"))
		blobOut.Close()
	}()

	dsName := spec.CreateValueSpecString("nbs", s.DBDir, "ds")
	// Run() will return when blobOut is closed.
	s.MustRun(main, []string{"--stdin", dsName})

	sp, err := spec.ForPath(dsName + ".value")
	assert.NoError(err)
	defer sp.Close()

	expected := types.NewBlob(bytes.NewBufferString("abcdef"))
	assert.True(expected.Equals(sp.GetValue()))

	ds := sp.GetDatabase().GetDataset("ds")
	meta := ds.Head().Get(datas.MetaField).(types.Struct)
	// The meta should only have a "date" field.
	metaDesc := types.TypeOf(meta).Desc.(types.StructDesc)
	assert.Equal(1, metaDesc.Len())
	assert.NotNil(metaDesc.Field("date"))
}

func (s *testSuite) TestImportFromFile() {
	assert := s.Assert()

	f, err := ioutil.TempFile("", "TestImportFromFile")
	assert.NoError(err)

	f.Write([]byte("abcdef"))
	f.Close()

	dsName := spec.CreateValueSpecString("nbs", s.DBDir, "ds")
	s.MustRun(main, []string{f.Name(), dsName})

	sp, err := spec.ForPath(dsName + ".value")
	assert.NoError(err)
	defer sp.Close()

	expected := types.NewBlob(bytes.NewBufferString("abcdef"))
	assert.True(expected.Equals(sp.GetValue()))

	ds := sp.GetDatabase().GetDataset("ds")
	meta := ds.Head().Get(datas.MetaField).(types.Struct)
	metaDesc := types.TypeOf(meta).Desc.(types.StructDesc)
	assert.Equal(2, metaDesc.Len())
	assert.NotNil(metaDesc.Field("date"))
	assert.Equal(f.Name(), string(meta.Get("file").(types.String)))
}

// TODO: TestImportFromURL
