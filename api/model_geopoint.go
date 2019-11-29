package api

import (
	"encoding/json"
	"io"

	"github.com/friendsofgo/errors"
)

type GeoPoint [2]float64

func NewGeoPoint(lat, lng float64) (g GeoPoint) {
	g[0] = lat
	g[1] = lng
	return
}

func (g GeoPoint) Lat() float64 {
	return g[0]
}

func (g GeoPoint) Lng() float64 {
	return g[1]
}

func (g GeoPoint) MarshalGQL(w io.Writer) {
	enc := json.NewEncoder(w)
	_ = enc.Encode(g)
}

func (g *GeoPoint) UnmarshalGQL(v interface{}) error {
	switch v := v.(type) {
	case []interface{}:
		if len(v) != 2 {
			return errors.New("GeoPoint must have 2 entries: [lat, lng]")
		}

		lat, ok := v[0].(json.Number)
		if !ok {
			return errors.Errorf("lat must be numeric, %T given", v[0])
		}
		lng, ok := v[1].(json.Number)
		if !ok {
			return errors.Errorf("lat must be numeric, %T given", v[1])
		}

		var err error
		g[0], err = lat.Float64()
		if err != nil {
			return errors.Wrap(err, "could not parse lat to float64")
		}
		g[1], err = lng.Float64()
		if err != nil {
			return errors.Wrap(err, "could not parse lng to float64")
		}
		return nil
	default:
		return errors.Errorf("%T is not a float array", v)
	}
}
