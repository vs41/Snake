#ifndef GAMEENGINE_H
#define GAMEENGINE_H

#include "GameBoard.h"

class GameEngine {
private:
    GameBoard board;
    bool isRunning;
    bool isPaused;
    int score;
    int highScore;
    int speed;

public:
    GameEngine(int width, int height);
    void run();
    void handleInput();
    void update();
    void checkCollisions();
    void loadHighScore();
    void saveHighScore();
};

#endif
