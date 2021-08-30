package com

import "github.com/SENERGY-Platform/import-repository/lib/config"

type Com struct {
	config config.Config
}

func New(config config.Config) *Com {
	return &Com{config: config}
}

