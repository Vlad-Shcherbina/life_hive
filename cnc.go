package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
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
var current_grid [][]int

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

func updateState(report StatusReport) {
	players = make(map[int]Player)
	for _, player := range report.Players {
		players[player.Id] = player
	}
	players[0] = Player{}
	if current_grid == nil {
		fmt.Println("First grid")
		current_grid = report.Cells
		return
	}

	diff_with_current := GridDiff(current_grid, report.Cells)
	diff_with_predicted := GridDiff(GridStep(current_grid), report.Cells)

	if len(diff_with_current) <= len(diff_with_predicted) {
		fmt.Println(diff_with_current)
	} else {
		fmt.Println("New generation")
		fmt.Println(diff_with_predicted)
	}

	current_grid = report.Cells
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
		Grid:    current_grid,
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
