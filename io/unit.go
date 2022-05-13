package io

type Unit struct {}

var Unit1 = Unit{}

var IOUnit1 = Lift(Unit1)

type IOUnit = IO[Unit]
