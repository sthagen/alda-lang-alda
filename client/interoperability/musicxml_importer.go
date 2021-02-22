package interoperability

import (
	"alda.io/client/model"
	"github.com/clbanning/mxj"
	"io"
)

func ImportMusicXML(r io.Reader) (*model.Score, error) {
	// To interface with MusicXML, we use MXJ, which parses XML into an ordered map[string]interface{}
	// To store any information outside the root, we must process the base level of the file individually
	// See https://github.com/clbanning/mxj/issues/17
	m := make(mxj.Map)

	for {
		nestedMap, err := mxj.NewMapXmlSeqReader(r)
		if err != nil {
			if err == io.EOF {
				break
			}
			if err != mxj.NoRoot {
				return nil, err
			}
		}

		for key, val := range nestedMap {
			m[key] = val
		}
	}

	return ImportMusicXMLFromMap(m)
}

func ImportMusicXMLFromMap(m mxj.Map) (*model.Score, error) {
	return nil, nil
}
