#include "GameBoard.h"
#include <iostream>

GameBoard::GameBoard(int w, int h) : width(w), height(h) {
    auto occupied = getOccupied();
    food.generate(occupied, width, height);
    obstacle.generate(10, occupied, width, height);
}

bool GameBoard::isInside(const Position& pos) const {
    return pos.x >= 0 && pos.y >= 0 && pos.x < width && pos.y < height;
}

std::set<Position> GameBoard::getOccupied() const {
    std::set<Position> occupied(snake.getBody().begin(), snake.getBody().end());
    return occupied;
}

void GameBoard::render() const {
    system("clear");
    for (int y = 0; y < height; ++y) {
        for (int x = 0; x < width; ++x) {
            Position p = {x, y};
            if (p == snake.getHead()) std::cout << "@";
            else if (snake.isColliding(p)) std::cout << "o";
            else if (p == food.getLocation()) std::cout << "*";
            else if (obstacle.isObstacle(p)) std::cout << "#";
            else std::cout << ".";
        }
        std::cout << "\n";
    }
}
