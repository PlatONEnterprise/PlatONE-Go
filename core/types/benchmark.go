package types

type Benchmark struct {
	To          string `json:"to" gencodec:"required"`
	Count       string `json:"count" gencodec:"required"`
	PreGenerate bool   `json:"preGenerate" gencodec:"required"`
	ProducerCnt string `json:"producerCnt" gencodec:"required"`
}
