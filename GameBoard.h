#ifndef GAMEBOARD_H
#define GAMEBOARD_H

#include "Snake.h"
#include "Food.h"
#include "Obstacle.h"
#include <set>

class GameBoard {
private:
    int width, height;

public:
    Snake snake;
    Food food;
    Obstacle obstacle;

    GameBoard(int w, int h);
    bool isInside(const Position& pos) const;
    void render() const;
    std::set<Position> getOccupied() const;
};

#endif
