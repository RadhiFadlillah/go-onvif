package onvif

// Device contains data of ONVIF camera
type Device struct {
	ID       string
	Name     string
	XAddr    string
	User     string
	Password string
}

// DeviceInformation contains information of ONVIF camera
type DeviceInformation struct {
	FirmwareVersion string
	HardwareID      string
	Manufacturer    string
	Model           string
	SerialNumber    string
}

// TimeZone data for GetSystemTimeAndDate
type TimeZone struct {
	TZ string
}

// UTCDateTime data for GetSystemTimeAndDate
type UTCDateTime struct {
	Time Time
	Date Date
}

// LocalDateTime data for GetSystemTimeAndDate
type LocalDateTime struct {
	Time Time
	Date Date
}

// Time data for GetSystemTimeAndDate
type Time struct {
	Hour   int
	Minute int
	Second int
}

// Date data for GetSystemTimeAndDate
type Date struct {
	Year  int
	Month int
	Day   int
}

// SystemDateAndTime data for GetSystemTimeAndDate
type SystemDateAndTime struct {
	DateTimeType    string
	DaylightSavings bool
	TimeZone        TimeZone
	UTCDateTime     UTCDateTime
	LocalDateTime   LocalDateTime
}

// NetworkCapabilities contains networking capabilities of ONVIF camera
type NetworkCapabilities struct {
	DynDNS     bool
	IPFilter   bool
	IPVersion6 bool
	ZeroConfig bool
}

// DeviceCapabilities contains capabilities of an ONVIF camera
type DeviceCapabilities struct {
	Network   NetworkCapabilities
	Events    map[string]bool
	Streaming map[string]bool
}

// HostnameInformation contains hostname info of an ONVIF camera
type HostnameInformation struct {
	Name     string
	FromDHCP bool
}

// MediaBounds contains resolution of a video media
type MediaBounds struct {
	Height int
	Width  int
}

// MediaSourceConfig contains configuration of a media source
type MediaSourceConfig struct {
	Name        string
	Token       string
	SourceToken string
	Bounds      MediaBounds
}

// VideoRateControl contains rate control of a video
type VideoRateControl struct {
	BitrateLimit     int
	EncodingInterval int
	FrameRateLimit   int
}

// VideoEncoderConfig contains configuration of a video encoder
type VideoEncoderConfig struct {
	Name           string
	Token          string
	Encoding       string
	Quality        int
	RateControl    VideoRateControl
	Resolution     MediaBounds
	SessionTimeout string
}

// AudioEncoderConfig contains configuration of an audio encoder
type AudioEncoderConfig struct {
	Name           string
	Token          string
	Encoding       string
	Bitrate        int
	SampleRate     int
	SessionTimeout string
}

// PTZConfig contains configuration of a PTZ control in camera
type PTZConfig struct {
	Name      string
	Token     string
	NodeToken string
}

// MediaProfile contains media profile of an ONVIF camera
type MediaProfile struct {
	Name               string
	Token              string
	VideoSourceConfig  MediaSourceConfig
	VideoEncoderConfig VideoEncoderConfig
	AudioSourceConfig  MediaSourceConfig
	AudioEncoderConfig AudioEncoderConfig
	PTZConfig          PTZConfig
}

// MediaURI contains streaming URI of an ONVIF camera
type MediaURI struct {
	URI                 string
	Timeout             string
	InvalidAfterConnect bool
	InvalidAfterReboot  bool
}
