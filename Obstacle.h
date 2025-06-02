#ifndef OBSTACLE_H
#define OBSTACLE_H

#include "Position.h"
#include <vector>
#include <set>

class Obstacle {
private:
    std::vector<Position> blocks;

public:
    void generate(int count, const std::set<Position>& occupied, int width, int height);
    bool isObstacle(const Position& pos) const;
    const std::vector<Position>& getBlocks() const;
};

#endif
