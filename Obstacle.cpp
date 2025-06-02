#include "Obstacle.h"
#include <cstdlib>
#include <ctime>

void Obstacle::generate(int count, const std::set<Position>& occupied, int width, int height) {
    srand(time(nullptr));
    while (blocks.size() < static_cast<size_t>(count)) {
        Position pos = { rand() % width, rand() % height };
        if (occupied.find(pos) == occupied.end()) {
            blocks.push_back(pos);
        }
    }
}

bool Obstacle::isObstacle(const Position& pos) const {
    for (const auto& b : blocks) {
        if (b == pos) return true;
    }
    return false;
}

const std::vector<Position>& Obstacle::getBlocks() const {
    return blocks;
}
