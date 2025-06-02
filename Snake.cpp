#include "Snake.h"

Snake::Snake() : dir(Direction::RIGHT) {
    body.push_back({5, 5});
}

void Snake::move() {
    Position head = getHead();
    switch (dir) {
        case Direction::UP:    head.y--; break;
        case Direction::DOWN:  head.y++; break;
        case Direction::LEFT:  head.x--; break;
        case Direction::RIGHT: head.x++; break;
    }
    body.push_front(head);
    body.pop_back();
}

void Snake::grow() {
    Position head = getHead();
    switch (dir) {
        case Direction::UP:    head.y--; break;
        case Direction::DOWN:  head.y++; break;
        case Direction::LEFT:  head.x--; break;
        case Direction::RIGHT: head.x++; break;
    }
    body.push_front(head);
}

void Snake::changeDirection(Direction newDir) {
    if ((dir == Direction::UP && newDir != Direction::DOWN) ||
        (dir == Direction::DOWN && newDir != Direction::UP) ||
        (dir == Direction::LEFT && newDir != Direction::RIGHT) ||
        (dir == Direction::RIGHT && newDir != Direction::LEFT)) {
        dir = newDir;
    }
}

bool Snake::isColliding(const Position& pos) const {
    for (const auto& segment : body) {
        if (segment == pos) return true;
    }
    return false;
}

Position Snake::getHead() const {
    return body.front();
}

const std::deque<Position>& Snake::getBody() const {
    return body;
}

Direction Snake::getDirection() const {
    return dir;
}
