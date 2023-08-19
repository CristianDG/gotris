// vim: foldmethod=marker
// vim: foldmarker={{{,}}}
package main

/*
 TODO:
 - [ ] fazer o tetris
   - [X] fazer todas as peças cairem
   - [ ] game over
   - [X] hold
     - [X] visualização
   - [X] next
     - [X] visualização
 - [ ] score
*/

import (
	"container/list"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Quantidade de pixels para um bloco
const BLOCK_SIZE = 20

const WINDOW_HEIGHT = 540
const WINDOW_WIDTH = 960

var background = rl.Black
var foreground = rl.RayWhite

type State uint8

const (
	UP State = iota
	LEFT
	DOWN
	RIGHT
)

type TetrominoEnum uint8

const (
	I TetrominoEnum = iota
	J
	L
	O
	T
	S
	Z
)

/*
essa função seta os bits dado um indice
00 01 02   03
04 05 06   07
08 09 10   11

12 13 14   15
*/
func setGrid(indexes [4]uint8) uint16 {
	var res uint16
	for _, index := range indexes {
		res |= 1 << (15 - index)
	}
	return res
}

// As rotações são em ordem UP, LEFT, DOWN, RIGHT
var ROTATION_TABLE = map[TetrominoEnum][4]uint16{
	I: {setGrid([4]uint8{4, 5, 6, 7}), setGrid([4]uint8{2, 6, 10, 14}), setGrid([4]uint8{8, 9, 10, 11}), setGrid([4]uint8{1, 5, 9, 13})},
	J: {setGrid([4]uint8{4, 5, 6, 10}), setGrid([4]uint8{1, 5, 8, 9}), setGrid([4]uint8{0, 4, 5, 6}), setGrid([4]uint8{1, 2, 5, 9})},
	L: {setGrid([4]uint8{2, 4, 5, 6}), setGrid([4]uint8{1, 5, 9, 10}), setGrid([4]uint8{4, 5, 6, 8}), setGrid([4]uint8{0, 1, 5, 9})},
	O: {setGrid([4]uint8{1, 2, 5, 6}), setGrid([4]uint8{1, 2, 5, 6}), setGrid([4]uint8{1, 2, 5, 6}), setGrid([4]uint8{1, 2, 5, 6})},
	T: {setGrid([4]uint8{1, 4, 5, 6}), setGrid([4]uint8{1, 5, 9, 6}), setGrid([4]uint8{4, 5, 6, 9}), setGrid([4]uint8{1, 4, 5, 9})},
	S: {setGrid([4]uint8{1, 2, 4, 5}), setGrid([4]uint8{1, 5, 6, 10}), setGrid([4]uint8{5, 6, 8, 9}), setGrid([4]uint8{0, 4, 5, 9})},
	Z: {setGrid([4]uint8{0, 1, 5, 6}), setGrid([4]uint8{2, 5, 6, 9}), setGrid([4]uint8{4, 5, 9, 10}), setGrid([4]uint8{1, 4, 5, 8})},
}

type Tetromino struct {
	piece TetrominoEnum
	state State
}

func (self Tetromino) isSet(x int, y int) bool {
	return (ROTATION_TABLE[self.piece][self.state] & (1 << (15 - (4*y + x)))) > 0
}

func (self Tetromino) rotatedLeft() Tetromino {
	self.state = map[State]State{
		UP:    RIGHT,
		RIGHT: DOWN,
		DOWN:  LEFT,
		LEFT:  UP,
	}[self.state]
	return self
}

func (self Tetromino) rotatedRight() Tetromino {
	self.state = map[State]State{
		UP:    LEFT,
		LEFT:  DOWN,
		DOWN:  RIGHT,
		RIGHT: UP,
	}[self.state]
	return self
}

func (self TetrominoEnum) width() int {
	var res int
	switch self {
	case I:
		res = 4
	case O:
		res = 2
	default:
		res = 3
	}
	return res
}

func (self TetrominoEnum) height() int {
	if self == I {
		return 1
	}
	return 2
}

func (self TetrominoEnum) index_map() [8]bool {
	setted := func(indexes ...int) [8]bool {
		res := [8]bool{}
		for _, i := range indexes {
			res[i] = true
		}
		return res
	}
	return map[TetrominoEnum]([8]bool){
		T: setted(1, 4, 5, 6),
		I: setted(0, 1, 2, 3),
		J: setted(0, 1, 2, 6),
		L: setted(2, 4, 5, 6),
		O: setted(0, 1, 4, 5),
		S: setted(1, 2, 4, 5),
		Z: setted(0, 1, 5, 6),
	}[self]
}

const board_cols = 10
const board_rows = 23

var BOARD [board_rows][board_cols]bool

func getRandomTetromino(previous *TetrominoEnum) TetrominoEnum {
	tetromino := TetrominoEnum(rl.GetRandomValue(int32(I), int32(Z)))
	if previous == nil || tetromino != *previous {
		return tetromino
	}
	return getRandomTetromino(previous)
}

func drawTetrominoInFrame(frame_x, frame_y, frame_w, frame_h int32, tetromino TetrominoEnum) {

	mid_x := frame_x + (frame_w / 2)
	mid_y := frame_y + (frame_h / 2)
	mid_hold_x := float32(tetromino.width()) / 2.0
	mid_hold_y := float32(tetromino.height()) / 2.0

	for i := 0; i < 2; i++ {
		for j := 0; j < 4; j++ {
			color := rl.Gray
			if tetromino.index_map()[i*4+j] {
				color = foreground

				var x = int32(j*BLOCK_SIZE) + (mid_x - int32(mid_hold_x*float32(BLOCK_SIZE)))
				var y = int32(i*BLOCK_SIZE) + (mid_y - int32(mid_hold_y*float32(BLOCK_SIZE)))
				rl.DrawRectangle(x, y, BLOCK_SIZE, BLOCK_SIZE, color)

			}
		}
	}
}

func canPlace(tetromino Tetromino, x int, y int) bool {

	var res = true

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			var can_place_x = (j+x) >= 0 && (j+x) < board_cols
			var can_place_y = (i+y) >= 0 && (i+y) < board_rows
			res = res && (!tetromino.isSet(j, i) || (can_place_x && can_place_y && (!BOARD[i+y][j+x])))

			if !res {
				return res
			}
		}
	}

	return res
}

var NEXT_TETROMINOS = list.New()

func main() {

	rl.InitWindow(WINDOW_WIDTH, WINDOW_HEIGHT, "Gotris")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	NEXT_TETROMINOS.PushBack(getRandomTetromino(nil))

	populateNext := func(){
		var previous TetrominoEnum = NEXT_TETROMINOS.Back().Value.(TetrominoEnum)
		NEXT_TETROMINOS.PushBack(getRandomTetromino(&previous))
	}

	for i := 0; i < 4; i++ {
		populateNext()
	}

	var tip = 0
	var tetrominos = []TetrominoEnum{I, J, L, O, T, S, Z}
	var tetromino = Tetromino{S, UP}

	var tetromino_x = 0
	var tetromino_y = 0
	var step_timer float32 = 0
	var place_tries = 0
	var can_hold = true
	var is_holding = false
	var hold TetrominoEnum
	var is_paused = false

	next := func() Tetromino {

		el := NEXT_TETROMINOS.Front()
		var piece TetrominoEnum = el.Value.(TetrominoEnum)
		var tetromino = Tetromino{piece: piece, state: UP}
		NEXT_TETROMINOS.Remove(el)
		populateNext()

		return tetromino
	}

	step := func() {
		if is_paused {
			return
		}
		if canPlace(tetromino, tetromino_x, tetromino_y+1) {
			tetromino_y += 1
		} else {
			place_tries += 1
			if place_tries == 3 {
				// hora de colocar o tetromino
				place_tries = 0
				can_hold = true

				// TODO: mudar para outra função
				for i := 0; i < 4; i++ {
					for j := 0; j < 4; j++ {
						if (i+tetromino_y >= 0 && i+tetromino_y < board_rows) && (j+tetromino_x >= 0 && j+tetromino_x < board_cols) && tetromino.isSet(j, i) {
							BOARD[i+tetromino_y][j+tetromino_x] = true
						}
					}
				}
				tetromino = next()
				tetromino_x = 0
				tetromino_y = 0
			}
		}
	}

	for {
		if rl.WindowShouldClose() || rl.IsKeyPressed(rl.KeyQ) {
			break
		}

		// TODO: controlar se aqui é o melhor lugar
		step_timer += rl.GetFrameTime()
		if step_timer >= 0.2 {
			step_timer = 0
			step()
		}

		// input {{{
		for {
			// TODO: trocar para ifs :+1: e usar keydown tbm

			var key = rl.GetKeyPressed()
			if key == 0 {
				break
			}
			switch key {
			case rl.KeyR, rl.KeyUp:
				{
					// NOTE: fds :+1:
					if canPlace(tetromino.rotatedRight(), tetromino_x, tetromino_y) {
						tetromino = tetromino.rotatedRight()
					} else if canPlace(tetromino.rotatedRight(), tetromino_x, tetromino_y-1) {
						// FIXME: isso não funciona para peças como L J e I
						tetromino_y -= 1
						tetromino = tetromino.rotatedRight()
					} else if canPlace(tetromino.rotatedRight(), tetromino_x+1, tetromino_y) {
						tetromino_x += 1
						tetromino = tetromino.rotatedRight()
					} else if canPlace(tetromino.rotatedRight(), tetromino_x-1, tetromino_y) {
						tetromino_x -= 1
						tetromino = tetromino.rotatedRight()
					}
				}
			case rl.KeyLeft, rl.KeyH:
				{
					if canPlace(tetromino, tetromino_x-1, tetromino_y) {
						tetromino_x -= 1
					}
				}
			case rl.KeyRight, rl.KeyL:
				{
					if canPlace(tetromino, tetromino_x+1, tetromino_y) {
						tetromino_x += 1
					}
				}
			case rl.KeyDown, rl.KeyJ:
				{
					step()
				}
			case rl.KeySpace:
				{
					for {
						if tetromino_y == 0 {
							break
						}
						step()
					}

				}

			case rl.KeyP:
				is_paused = !is_paused
			case rl.KeyC:
				{
					if !can_hold {
						break
					}

					can_hold = false

					temp := hold
					hold = tetromino.piece

					if !is_holding {
						is_holding = true
						tetromino = next()
					} else {
						tetromino = Tetromino{temp, UP}
					}

					tetromino_x = 0
					tetromino_y = 0
				}
			// NOTE: só para debug
			case rl.KeyE:
				{
					tetromino_x = 0
					tetromino_y = 0
					tip = (tip + 1) % len(tetrominos)
					tetromino.piece = tetrominos[tip]
				}

			}

		}
		// }}} input

		// detectando se uma ou mais linhas foram preenchidas {{{
		var isFilled = func(line [10]bool) bool {
			var res = true
			for _, block := range line {
				res = res && block
			}
			return res
		}

		var lines_filled []int
		for index, line := range BOARD {
			if isFilled(line) {
				lines_filled = append(lines_filled, index)
			}
		}

		for _, index := range lines_filled {
			for i := index; i > 0; i-- {
				BOARD[i] = BOARD[i-1]
			}
			BOARD[0] = [10]bool{}
		}
		// }}}

		rl.BeginDrawing()

		rl.ClearBackground(background)

		// board_frame - tamanho interno de 10 blocos de largura e 20 de altura {{{

		var board_frame_width int32 = board_cols * BLOCK_SIZE
		var board_frame_height int32 = board_rows * BLOCK_SIZE

		var board_frame_start_pos_x int32 = (WINDOW_WIDTH / 2) - (board_frame_width / 2)
		var board_frame_start_pos_y int32 = (WINDOW_HEIGHT / 2) - (board_frame_height / 2)

		var board_frame_end_pos_x int32 = board_frame_start_pos_x + board_frame_width
		var board_frame_end_pos_y int32 = board_frame_start_pos_y + board_frame_height

		// > board_frame wireframe
		drawVLine := func(offset int, color rl.Color) {
			rl.DrawLine(board_frame_start_pos_x+int32(offset*BLOCK_SIZE), board_frame_start_pos_y, board_frame_start_pos_x+int32(offset*BLOCK_SIZE), board_frame_end_pos_y, color)
		}

		drawHLine := func(offset int, color rl.Color) {
			rl.DrawLine(board_frame_start_pos_x, board_frame_start_pos_y+int32(offset*BLOCK_SIZE), board_frame_end_pos_x, board_frame_start_pos_y+int32(offset*BLOCK_SIZE), color)
		}

		grid_color := rl.Color{32, 32, 32, 255}
		drawVLine(0, foreground)          // parede esquerda
		drawVLine(board_cols, foreground) // parede direita
		for i := 1; i < board_cols; i++ {
			drawVLine(i, grid_color)
		}

		drawHLine(board_rows, foreground) // fundo
		for i := 1; i < board_rows; i++ {
			drawHLine(i, grid_color)
		}

		// }}} board_frame

		// hold_frame - tamanho interno 10 por 4 {{{

		var hold_frame_width int32 = 6 * BLOCK_SIZE
		var hold_frame_height int32 = 4 * BLOCK_SIZE

		var hold_frame_start_pos_x int32 = board_frame_start_pos_x - hold_frame_width - BLOCK_SIZE
		var hold_frame_start_pos_y int32 = board_frame_start_pos_y + BLOCK_SIZE

		rl.DrawText("hold", hold_frame_start_pos_x+hold_frame_width-int32(len("hold")*10), hold_frame_start_pos_y-BLOCK_SIZE, 20, foreground)
		rl.DrawRectangleLines(hold_frame_start_pos_x, hold_frame_start_pos_y, hold_frame_width, hold_frame_height, foreground)

		// TODO: desenhar um tetromino dentro do display de hold

		if is_holding {
			drawTetrominoInFrame(hold_frame_start_pos_x, hold_frame_start_pos_y, hold_frame_width, hold_frame_height, hold)
		}

		// }}} hold_frame

		// next_frame - tamanho interno de 5 peças por 1 {{{

		var next_frame_width int32 = 6 * BLOCK_SIZE
		var next_single_frame_height int32 = 4 * BLOCK_SIZE
		var next_frame_height int32 = next_single_frame_height * int32(NEXT_TETROMINOS.Len())

		var next_frame_start_pos_x int32 = board_frame_end_pos_x + BLOCK_SIZE
		var next_frame_start_pos_y int32 = board_frame_start_pos_y + BLOCK_SIZE

		rl.DrawText("next", next_frame_start_pos_x, next_frame_start_pos_y-BLOCK_SIZE, 20, foreground)
		rl.DrawRectangleLines(next_frame_start_pos_x, next_frame_start_pos_y, next_frame_width, next_frame_height, foreground)

		var i = 0
		for e := NEXT_TETROMINOS.Front(); e != nil; e = e.Next() {
			display_item := e.Value.(TetrominoEnum)

			drawTetrominoInFrame(
				next_frame_start_pos_x,
				next_frame_start_pos_y+int32(i)*next_single_frame_height,
				next_frame_width,
				next_single_frame_height,
				display_item)

			i++
		}

		// }}} next_frame

		// desenhando board
		for i := 0; i < int(board_rows); i++ {
			for j := 0; j < int(board_cols); j++ {
				if BOARD[i][j] {
					var x = int32(j*BLOCK_SIZE) + board_frame_start_pos_x
					var y = int32(i*BLOCK_SIZE) + board_frame_start_pos_y
					rl.DrawRectangle(x, y, BLOCK_SIZE, BLOCK_SIZE, rl.LightGray)
				}
			}
		}

		// desenhando o tetromino
		for i := 0; i < 4; i++ {
			for j := 0; j < 4; j++ {
				if tetromino.isSet(j, i) {
					var x = int32(j*BLOCK_SIZE) + board_frame_start_pos_x + int32(tetromino_x*BLOCK_SIZE)
					var y = int32(i*BLOCK_SIZE) + board_frame_start_pos_y + int32(tetromino_y*BLOCK_SIZE)
					rl.DrawRectangle(x, y, BLOCK_SIZE, BLOCK_SIZE, foreground)
				}
			}
		}


		rl.EndDrawing()
	}
}
