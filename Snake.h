#ifndef SNAKE_H
#define SNAKE_H

#include <deque>
#include "Position.h"
#include "Direction.h"

class Snake {
private:
    std::deque<Position> body;
    Direction dir;

public:
    Snake();
    void move();
    void grow();
    void changeDirection(Direction newDir);
    bool isColliding(const Position& pos) const;
    Position getHead() const;
    const std::deque<Position>& getBody() const;
    Direction getDirection() const;
};

#endif
