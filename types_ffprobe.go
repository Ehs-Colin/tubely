package main

type FfprobeFile struct {
	Streams  []*Stream  `json:"streams"`
	Chapters []*Chapter `json:"chapters"`
	Format   *Format    `json:"format"`
	// contains filtered or unexported fields
}

type Fraction string

type Stream struct {
	Index          int      `json:"index"` // stream index
	CodecName      string   `json:"codec_name"`
	CodecLongName  string   `json:"codec_long_name"`
	Profile        string   `json:"profile"`
	CodecType      string   `json:"codec_type"`       // video
	CodecTagString string   `json:"codec_tag_string"` // avc1
	CodecTag       string   `json:"codec_tag"`        // 0x31637661
	Width          int      `json:"width,omitempty"`
	Height         int      `json:"height,omitempty"`
	CodedWidth     int      `json:"coded_width"`
	CodedHeight    int      `json:"coded_height"`
	ClosedCaptions int      `json:"closed_captions"`
	FrameRate      Fraction `json:"r_frame_rate"`
	AvgFrameRate   Fraction `json:"avg_frame_rate"`
	TimeBase       Fraction `json:"time_base"`
	SampleRate     int      `json:"sample_rate,string"`
	Duration       float64  `json:"duration,string"`
	DurationTs     uint64   `json:"duration_ts"`
	StartTime      float64  `json:"start_time,string"` // can be negative
	StartPTS       int      `json:"start_pts"`
	ExtradataSize  int      `json:"extradata_size"`

	Disposition *struct {
		Default         int `json:"default"`
		Dub             int `json:"dub"`
		Original        int `json:"original"`
		Comment         int `json:"comment"`
		Lyrics          int `json:"lyrics"`
		Karaoke         int `json:"karaoke"`
		Forced          int `json:"forced"`
		HearingImpaired int `json:"hearing_impaired"`
		VisualImpaired  int `json:"visual_impaired"`
		CleanEffects    int `json:"clean_effects"`
		AttachedPic     int `json:"attached_pic"`
		TimedThumbnails int `json:"timed_thumbnails"`
		Captions        int `json:"captions"`
		Descriptions    int `json:"descriptions"`
		Metadata        int `json:"metadata"`
		Dependent       int `json:"dependent"`
		StillImage      int `json:"still_image"`
	} `json:"disposition,omitempty"`

	Tags map[string]string `json:"tags"`
}

type Chapter struct {
	Id        int64             `json:"id"`
	TimeBase  Fraction          `json:"time_base"`
	Start     int64             `json:"start"`
	StartTime float64           `json:"start_time,string"`
	End       int64             `json:"end"`
	EndTime   float64           `json:"end_time,string"`
	Tags      map[string]string `json:"tags,omitempty"`
}

type Format struct {
	Filename       string            `json:"filename"`
	FormatName     string            `json:"format_name"`       // mov,mp4,m4a,3gp,3g2,mj2
	FormatLongName string            `json:"format_long_name"`  // QuickTime / MOV
	StartTime      float64           `json:"start_time,string"` // "start_time": "0.000000",
	Duration       float64           `json:"duration,string"`   // "duration": "240.048000",
	Size           int64             `json:"size,string"`       // "size": "73087904",
	BitRate        int               `json:"bit_rate,string"`   // "bit_rate": "2435776",
	Tags           map[string]string `json:"tags"`
}
