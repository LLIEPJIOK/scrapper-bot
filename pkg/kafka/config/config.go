package config

import "time"

type Kafka struct {
	Producer Producer `envPrefix:"PRODUCER_"`
	Consumer Consumer `envPrefix:"CONSUMER_"`
	Retrier  Retrier  `envPrefix:"RETRY_"`
	DLQ      DLQ      `envPrefix:"DLQ_"`
	Brokers  []string `                      env:"BROKERS,required"`
}

type Producer struct {
	RetryMax        int    `env:"RETRY_MAX"        envDefault:"3"`
	ReturnSuccesses bool   `env:"RETURN_SUCCESSES" envDefault:"true"`
	ReturnErrors    bool   `env:"RETURN_ERRORS"    envDefault:"true"`
	RequiredAcks    int16  `env:"REQUIRED_ACKS"    envDefault:"1"`
	Partitioner     string `env:"PARTITIONER"      envDefault:"random"`
}

type Consumer struct {
	Topics       []string `env:"TOPICS,required"`
	ReturnErrors bool     `env:"RETURN_ERRORS"   envDefault:"true"`
	Group        string   `env:"GROUP"           envDefault:"default"`
}

type Retrier struct {
	MaxRetries    int32         `env:"MAX_RETRIES"    envDefault:"3"`
	TableName     string        `env:"TABLE_NAME"     envDefault:"kafka_retrier"`
	CheckInterval time.Duration `env:"CHECK_INTERVAL" envDefault:"10m"`
	InitialDelay  time.Duration `env:"INITIAL_DELAY"  envDefault:"10m"`
}

type DLQ struct {
	TableName string `env:"TABLE_NAME" envDefault:"kafka_dlq"`
}
