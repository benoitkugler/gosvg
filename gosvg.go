package gosvg

import (
	"fmt"
	"io"

	"github.com/benoitkugler/webrender/svg"
)

func Render(src io.Reader) error {
	icon, err := svg.Parse(src, "", nil, nil)
	if err != nil {
		return err
	}

	output := NewCanvas(0, 0, 600, 600)
	icon.Draw(output, 600, 600, nil)

	fmt.Println(output.path)

	return nil
}
