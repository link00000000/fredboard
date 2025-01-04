package content

import "embed"

//go:embed static/* templates/*
var ContentFS embed.FS
