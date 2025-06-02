#ifndef FOOD_H
#define FOOD_H

#include "Position.h"
#include <set>

class Food {
private:
    Position location;

public:
    void generate(const std::set<Position>& occupied, int width, int height);
    Position getLocation() const;
};

#endif
