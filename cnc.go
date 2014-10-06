package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"
	"os"
)

type Player struct {
	Id     int
	Name   string
	Color  string
	Online bool
}

type StatusReport struct {
	MyId    int
	Cells   [][]int
	Players []Player
}

type Command struct {
	Debug string
}

var players map[int]Player
var currentGrid [][]int
var currentGeneration int = 0

func GridSize(grid [][]int) (width, height int) {
	height = len(grid)
	width = len(grid[0])
	return
}

func GetNeighbors(grid [][]int, i, j int, result *[8]int) {
	w, h := GridSize(grid)
	q := 0
	for ii := i + h - 1; ii <= i+h+1; ii++ {
		for jj := j + w - 1; jj <= j+w+1; jj++ {
			if ii == i+h && jj == j+w {
				continue
			}
			result[q] = grid[ii%h][jj%w]
			q++
		}
	}
	if q != 8 {
		panic(q)
	}
	return
}

func GetDominantNeighbor(neighbors *[8]int) int {
	counts := map[int]int{}
	max_count := 0
	best_id := -1

	for _, x := range neighbors {
		if x != 0 {
			c, _ := counts[x]
			c++
			counts[x] = c
			if c > max_count {
				max_count = c
				best_id = x
			}
		}
	}
	if best_id == -1 {
		panic(best_id)
	}
	return best_id
}

func GridStep(grid [][]int) (new_grid [][]int) {
	width, height := GridSize(grid)
	new_grid = make([][]int, height)
	var neighbors [8]int
	for i := range new_grid {
		new_grid[i] = make([]int, width)
		for j := range new_grid[i] {
			GetNeighbors(grid, i, j, &neighbors)
			num_neighbors := 0
			for _, x := range neighbors {
				if x != 0 {
					num_neighbors++
				}
			}

			switch num_neighbors {
			case 2:
				new_grid[i][j] = grid[i][j]
			case 3:
				if grid[i][j] == 0 {
					new_grid[i][j] = GetDominantNeighbor(&neighbors)
				} else {
					new_grid[i][j] = grid[i][j]
				}
			default:
				new_grid[i][j] = 0
			}
		}
	}
	return
}

type DiffEntry struct {
	X, Y     int
	Old, New int
}

func GridDiff(old_grid, new_grid [][]int) []DiffEntry {
	var result []DiffEntry = nil
	for i, row := range new_grid {
		for j, cell := range row {
			if old_grid[i][j] != cell {
				result = append(result, DiffEntry{
					X: j, Y: i,
					Old: old_grid[i][j],
					New: cell})
			}
		}
	}
	return result
}

const NO_CHANGE = -1

type Patch struct {
	X, Y int
	Data [][]int
}

func EmptyPatch(x1, y1, x2, y2 int) Patch {
	new_cells := make([][]int, y2-y1)
	for i := range new_cells {
		new_cells[i] = make([]int, x2-x1)
		for j := range new_cells[i] {
			new_cells[i][j] = NO_CHANGE
		}
	}
	return Patch{X: x1, Y: y1, Data: new_cells}
}

func (patch *Patch) Resize(x1, y1, x2, y2 int) Patch {
	result := EmptyPatch(x1, y1, x2, y2)
	for i, row := range patch.Data {
		y := patch.Y + i
		for j, cell := range row {
			x := patch.X + j
			if x1 <= x && x < x2 && y1 <= y && y < y2 {
				result.Data[y-y1][x-x1] = cell
			} else {
				if cell != NO_CHANGE {
					panic(cell)
				}
			}
		}
	}
	return result
}

func (patch *Patch) Shrink() Patch {
	const inf = 1000000
	minX := inf
	maxX := -inf
	minY := inf
	maxY := -inf
	for i, row := range patch.Data {
		y := patch.Y + i
		for j, cell := range row {
			if cell != NO_CHANGE {
				x := patch.X + j
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if minX == inf {
		return Patch{X: patch.X, Y: patch.Y}
	}
	return patch.Resize(minX, minY, maxX+1, maxY+1)
}

func CssColorToRGB(color string) (r, g, b uint8) {
	if color[0] != '#' {
		panic(color)
	}
	bytes, err := hex.DecodeString(color[1:])
	if err != nil {
		panic(err)
	}
	if len(bytes) != 3 {
		panic(bytes)
	}
	r = bytes[0]
	g = bytes[1]
	b = bytes[2]
	return
}

func savePic() {
	const cellSize = 6
	im := image.NewRGBA(image.Rect(0, 0, 640, 360))

	bg := color.RGBA{128, 128, 128, 255}
	draw.Draw(im, im.Bounds(), &image.Uniform{bg}, image.ZP, draw.Src)

	for y, row := range currentGrid {
		for x, cell := range row {
			var color_ color.NRGBA
			if cell == 0 {
				color_ = color.NRGBA{255, 255, 255, 255}
			} else {
				r, g, b := CssColorToRGB(players[cell].Color)
				color_ = color.NRGBA{r, g, b, 255}
			}
			for i := 0; i < cellSize; i++ {
				for j := 0; j < cellSize; j++ {
					im.Set(x*cellSize+j+20, y*cellSize+i+20, color_)
				}
			}
		}
	}
	fmt.Println(im.Bounds().Max.Y)

	fout, err := os.OpenFile(
		fmt.Sprintf("pics/frame%06d.png", currentGeneration),
		os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	err = png.Encode(fout, im)
	if err != nil {
		panic(err)
	}
}

func updateState(report StatusReport) {
	players = make(map[int]Player)
	for _, player := range report.Players {
		players[player.Id] = player
	}
	players[0] = Player{}
	if currentGrid == nil {
		fmt.Println("First grid")
		currentGrid = report.Cells
		savePic()
		return
	}

	diff_with_current := GridDiff(currentGrid, report.Cells)
	diff_with_predicted := GridDiff(GridStep(currentGrid), report.Cells)

	currentGrid = report.Cells

	if len(diff_with_current) <= len(diff_with_predicted) {
		fmt.Println(diff_with_current)
	} else {
		currentGeneration++
		fmt.Println("New generation", currentGeneration)
		fmt.Println(diff_with_predicted)
		savePic()
	}
}

func botHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	// fmt.Println("Handling request", req)
	if req.Method == "OPTIONS" {
		return
	}

	decoder := json.NewDecoder(req.Body)
	var report StatusReport
	err := decoder.Decode(&report)
	if err != nil {
		fmt.Println("Error parsing json:", err)
	}

	// fmt.Println("report:", report)

	var command Command
	command.Debug = fmt.Sprintf(
		"id:%d grid:%dx%d",
		report.MyId, len(report.Cells), len(report.Cells[0]))
	response, err := json.Marshal(command)
	if err != nil {
		fmt.Println("Error serializing json:", err)
	}
	updateState(report)
	fmt.Fprint(w, string(response))
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("templates/index.html")
	data := struct {
		Players map[int]Player
		Grid    [][]int
	}{
		Players: players,
		Grid:    currentGrid,
	}
	t.Execute(w, data)
}

const host = "localhost:8000"

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/bot", botHandler)
	fmt.Printf("Serving at http://%s\n", host)
	http.ListenAndServe(host, nil)
}
