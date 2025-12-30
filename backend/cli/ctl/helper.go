package main

import "github.com/urfave/cli/v2"

func flattenFlags(flagSlices ...[]cli.Flag) []cli.Flag {
	var result []cli.Flag
	for _, slice := range flagSlices {
		result = append(result, slice...)
	}

	return result
}
