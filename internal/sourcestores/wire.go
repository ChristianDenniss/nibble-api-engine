package sourcestores

import sourceentity "github.com/ChristianDenniss/go-data-model/source/entity"

type storeWire struct {
	ID              string       `json:"id"`
	ChannelID       string       `json:"channelId"`
	ExternalStoreID string       `json:"externalStoreId"`
	Name            string       `json:"name"`
	Location        locationWire `json:"location"`
	Phone           string       `json:"phone"`
}

type locationWire struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
}

func toStoreList(in []sourceentity.Store) []storeWire {
	out := make([]storeWire, len(in))
	for i, st := range in {
		out[i] = storeWire{
			ID: st.ID, ChannelID: st.ChannelID, ExternalStoreID: st.ExternalStoreID,
			Name: st.Name, Phone: st.Phone,
			Location: locationWire{
				Latitude: st.Location.Latitude, Longitude: st.Location.Longitude, Address: st.Location.Address,
			},
		}
	}
	return out
}
