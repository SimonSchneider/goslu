# Go Standard Library Utils

Extensions and helpers for the go stdlib. Targeted at common web app usages. Important is that it's zero dependencies.

## config

read config from flags and env vars (based on reflection). Minimal setup required.

## date

Counterpart to `time` but for dates with the lowest resolution being day.

## email

Very simple wrappers around email APIs.

## migrate

Basic SQL DB migration support.

## sid

Random ID generation, no need for UUID package to generate a random ID. Uses crypto to generate random numbers.

## sqlu

Utils around sql databases, currently only for `NullStr`.

## srvu

Http package utils. Mainly for server and handlers. Extended to also improve `http.Client`.

## templ

Utilities for working with templates. Currently mainly aimed at server side rendering and development to support "hot reload" 
vs precompiling and embedding in production.

