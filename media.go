package onvif

var mediaXMLNs = []string{
	`xmlns:trt="http://www.onvif.org/ver10/media/wsdl"`,
	`xmlns:tt="http://www.onvif.org/ver10/schema"`,
}

// GetProfiles fetch available media profiles of ONVIF camera
func (device Device) GetProfiles() ([]MediaProfile, error) {
	// Create SOAP
	soap := SOAP{
		Body:     "<trt:GetProfiles/>",
		XMLNs:    mediaXMLNs,
		User:     device.User,
		Password: device.Password,
	}

	// Send SOAP request
	response, err := soap.SendRequest(device.XAddr)
	if err != nil {
		return []MediaProfile{}, err
	}

	// Get and parse list of profile to interface
	ifaceProfiles, err := response.ValuesForPath("Envelope.Body.GetProfilesResponse.Profiles")
	if err != nil {
		return []MediaProfile{}, err
	}

	// Create initial result
	result := []MediaProfile{}

	// Parse each available profile
	for _, ifaceProfile := range ifaceProfiles {
		if mapProfile, ok := ifaceProfile.(map[string]interface{}); ok {
			// Parse name and token
			profile := MediaProfile{}
			profile.Name = interfaceToString(mapProfile["Name"])
			profile.Token = interfaceToString(mapProfile["-token"])

			// Parse video source configuration
			videoSource := MediaSourceConfig{}
			if mapVideoSource, ok := mapProfile["VideoSourceConfiguration"].(map[string]interface{}); ok {
				videoSource.Name = interfaceToString(mapVideoSource["Name"])
				videoSource.Token = interfaceToString(mapVideoSource["-token"])
				videoSource.SourceToken = interfaceToString(mapVideoSource["SourceToken"])

				// Parse video bounds
				bounds := MediaBounds{}
				if mapVideoBounds, ok := mapVideoSource["Bounds"].(map[string]interface{}); ok {
					bounds.Height = interfaceToInt(mapVideoBounds["-height"])
					bounds.Width = interfaceToInt(mapVideoBounds["-width"])
				}
				videoSource.Bounds = bounds
			}
			profile.VideoSourceConfig = videoSource

			// Parse video encoder configuration
			videoEncoder := VideoEncoderConfig{}
			if mapVideoEncoder, ok := mapProfile["VideoEncoderConfiguration"].(map[string]interface{}); ok {
				videoEncoder.Name = interfaceToString(mapVideoEncoder["Name"])
				videoEncoder.Token = interfaceToString(mapVideoEncoder["-token"])
				videoEncoder.Encoding = interfaceToString(mapVideoEncoder["Encoding"])
				videoEncoder.Quality = interfaceToInt(mapVideoEncoder["Quality"])
				videoEncoder.SessionTimeout = interfaceToString(mapVideoEncoder["SessionTimeout"])

				// Parse video rate control
				rateControl := VideoRateControl{}
				if mapVideoRate, ok := mapVideoEncoder["RateControl"].(map[string]interface{}); ok {
					rateControl.BitrateLimit = interfaceToInt(mapVideoRate["BitrateLimit"])
					rateControl.EncodingInterval = interfaceToInt(mapVideoRate["EncodingInterval"])
					rateControl.FrameRateLimit = interfaceToInt(mapVideoRate["FrameRateLimit"])
				}
				videoEncoder.RateControl = rateControl

				// Parse video resolution
				resolution := MediaBounds{}
				if mapVideoRes, ok := mapVideoEncoder["Resolution"].(map[string]interface{}); ok {
					resolution.Height = interfaceToInt(mapVideoRes["Height"])
					resolution.Width = interfaceToInt(mapVideoRes["Width"])
				}
				videoEncoder.Resolution = resolution
			}
			profile.VideoEncoderConfig = videoEncoder

			// Parse audio source configuration
			audioSource := MediaSourceConfig{}
			if mapAudioSource, ok := mapProfile["AudioSourceConfiguration"].(map[string]interface{}); ok {
				audioSource.Name = interfaceToString(mapAudioSource["Name"])
				audioSource.Token = interfaceToString(mapAudioSource["-token"])
				audioSource.SourceToken = interfaceToString(mapAudioSource["SourceToken"])
			}
			profile.AudioSourceConfig = audioSource

			// Parse audio encoder configuration
			audioEncoder := AudioEncoderConfig{}
			if mapAudioEncoder, ok := mapProfile["AudioEncoderConfiguration"].(map[string]interface{}); ok {
				audioEncoder.Name = interfaceToString(mapAudioEncoder["Name"])
				audioEncoder.Token = interfaceToString(mapAudioEncoder["-token"])
				audioEncoder.Encoding = interfaceToString(mapAudioEncoder["Encoding"])
				audioEncoder.Bitrate = interfaceToInt(mapAudioEncoder["Bitrate"])
				audioEncoder.SampleRate = interfaceToInt(mapAudioEncoder["SampleRate"])
				audioEncoder.SessionTimeout = interfaceToString(mapAudioEncoder["SessionTimeout"])
			}
			profile.AudioEncoderConfig = audioEncoder

			// Parse PTZ configuration
			ptzConfig := PTZConfig{}
			if mapPTZ, ok := mapProfile["PTZConfiguration"].(map[string]interface{}); ok {
				ptzConfig.Name = interfaceToString(mapPTZ["Name"])
				ptzConfig.Token = interfaceToString(mapPTZ["-token"])
				ptzConfig.NodeToken = interfaceToString(mapPTZ["NodeToken"])
			}
			profile.PTZConfig = ptzConfig

			// Push profile to result
			result = append(result, profile)
		}
	}

	return result, nil
}

// GetStreamURI fetch stream URI of a media profile.
// Possible protocol is UDP, HTTP or RTSP
func (device Device) GetStreamURI(profileToken, protocol string) (MediaURI, error) {
	// Create SOAP
	soap := SOAP{
		XMLNs: mediaXMLNs,
		Body: `<trt:GetStreamUri>
			<trt:StreamSetup>
				<tt:Stream>RTP-Unicast</tt:Stream>
				<tt:Transport><tt:Protocol>` + protocol + `</tt:Protocol></tt:Transport>
			</trt:StreamSetup>
			<trt:ProfileToken>` + profileToken + `</trt:ProfileToken>
		</trt:GetStreamUri>`,
		User:     device.User,
		Password: device.Password,
	}

	// Send SOAP request
	response, err := soap.SendRequest(device.XAddr)
	if err != nil {
		return MediaURI{}, err
	}

	// Parse response to interface
	ifaceURI, err := response.ValueForPath("Envelope.Body.GetStreamUriResponse.MediaUri")
	if err != nil {
		return MediaURI{}, err
	}

	// Parse interface to struct
	streamURI := MediaURI{}
	if mapURI, ok := ifaceURI.(map[string]interface{}); ok {
		streamURI.URI = interfaceToString(mapURI["Uri"])
		streamURI.Timeout = interfaceToString(mapURI["Timeout"])
		streamURI.InvalidAfterConnect = interfaceToBool(mapURI["InvalidAfterConnect"])
		streamURI.InvalidAfterReboot = interfaceToBool(mapURI["InvalidAfterReboot"])
	}

	return streamURI, nil
}
