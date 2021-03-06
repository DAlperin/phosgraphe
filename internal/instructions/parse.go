package instructions

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

type Instruction = string

type instructionStep struct {
	Instructions []Instruction
}

type Instructions []instructionStep

func (i *Instructions) String() string {
	var segments []string
	for _, step := range *i {
		var segment []string
		for _, instruction := range step.Instructions {
			segment = append(segment, instruction)
		}
		segments = append(segments, strings.Join(segment, ","))
	}
	return strings.Join(segments, "/")
}

func Parse(input string) Instructions {
	instructions := Instructions{}
	components := strings.Split(input, "/")
	for _, s := range components {
		step := instructionStep{}
		stepComponents := strings.Split(s, ",")
		for _, k := range stepComponents {
			step.Instructions = append(step.Instructions, k)
		}
		instructions = append(instructions, step)
	}
	return instructions
}

func Hash(instructions Instructions) string {
	h := sha256.New()
	var hashes []string

	for _, step := range instructions {
		parts := step.Instructions

		sort.Strings(parts)
		joined := strings.Join(parts, "")
		h.Write([]byte(joined))

		hash := hex.EncodeToString(h.Sum(nil))
		hashes = append(hashes, hash)
	}

	h.Reset()
	h.Write([]byte(strings.Join(hashes, "")))

	return hex.EncodeToString(h.Sum(nil))
}
