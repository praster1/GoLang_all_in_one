package main

import "fmt"

type Point struct{ X, Y int }

func movePoint(point *Point, dx, dy int) bool {
	if point == nil || point.X < -1000 || point.X > 1000 || point.Y < -1000 || point.Y > 1000 {
		return false
	}
	if dx < -1000 || dx > 1000 || dy < -1000 || dy > 1000 {
		return false
	}
	nextX, nextY := point.X+dx, point.Y+dy
	if nextX < -1000 || nextX > 1000 || nextY < -1000 || nextY > 1000 {
		return false
	}
	point.X, point.Y = nextX, nextY
	return true
}

func main() {
	point := Point{X: 10, Y: 20}
	ok := movePoint(&point, -3, 5)
	fmt.Printf("이동: %t, 좌표: (%d, %d)\n", ok, point.X, point.Y)
	ok = movePoint(&point, 1000, 0)
	fmt.Printf("범위 초과: %t, 좌표: (%d, %d)\n", ok, point.X, point.Y)
}
