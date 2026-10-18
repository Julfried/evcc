package charger

import (
	"time"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/util/sponsor"
	"github.com/lorenzodonini/ocpp-go/ocpp1.6/smartcharging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (suite *ocppTestSuite) TestCurrentControl() {
	old := sponsor.Subject
	sponsor.Subject = "test"
	suite.T().Cleanup(func() { sponsor.Subject = old })

	for _, tc := range []struct {
		name      string
		milliamps bool
	}{
		{"default", true},
		{"milliamps", true},
		{"whole-amps", false},
	} {
		suite.Run(tc.name, func() {
			t := suite.T()
			id := suite.stationID("current-" + tc.name)
			cp, _, _ := suite.startChargePoint(id, 1)
			h := &ocppProfileHandler{profiles: make(chan *smartcharging.SetChargingProfileRequest, 1)}
			cp.SetSmartChargingHandler(h)
			require.NoError(t, cp.Start(ocppTestUrl))

			config := map[string]any{"stationid": id, "connecttimeout": ocppTestConnectTimeout.String()}
			if tc.name != "default" {
				config["milliamps"] = tc.milliamps
			}
			c, err := NewFromConfig(t.Context(), "ocpp", config)
			require.NoError(t, err)

			ex, ok := api.Cap[api.ChargerEx](c)
			require.Equal(t, tc.milliamps, ok)
			wantCurrent := 7.0
			if ok {
				require.NoError(t, ex.MaxCurrentMillis(7.25))
				wantCurrent = 7.2
			} else {
				require.NoError(t, c.MaxCurrent(7))
			}

			select {
			case req := <-h.profiles:
				assert.Equal(t, wantCurrent, req.ChargingProfile.ChargingSchedule.ChargingSchedulePeriod[0].Limit)
			case <-time.After(time.Second):
				t.Fatal("missing charging profile")
			}
		})
	}
}
