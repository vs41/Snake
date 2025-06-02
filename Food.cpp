#include "Food.h"
#include <cstdlib>
#include <ctime>

void Food::generate(const std::set<Position>& occupied, int width, int height) {
    srand(time(nullptr));
    do {
        location = { rand() % width, rand() % height };
    } while (occupied.find(location) != occupied.end());
}

Position Food::getLocation() const {
    return location;
}
