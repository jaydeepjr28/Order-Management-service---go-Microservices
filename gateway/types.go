package main

import pb "github.com/jaydeepjr28/commons/api"

type CreateOrderRequest struct {
	Order		*pb.Order	`json:"order"`
	RedirectToURL	string		`json:"redirectToURL"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}
