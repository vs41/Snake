#include "GameEngine.h"
#include <iostream>
#include <fstream>
#include <thread>
#include <chrono>
#include <termios.h>
#include <unistd.h>
#include <fcntl.h>

GameEngine::GameEngine(int width, int height)
    : board(width, height), isRunning(true), isPaused(false), score(0), speed(200) {
    loadHighScore();
}

void GameEngine::run() {
    while (isRunning) {
        handleInput();
        if (!isPaused) {
            update();
            checkCollisions();
            board.render();
            std::cout << "Score: " << score << " High Score: " << highScore << "\n";
            std::this_thread::sleep_for(std::chrono::milliseconds(speed));
        }
    }
    saveHighScore();
    std::cout << "Game Over!\n";
}

void GameEngine::handleInput() {
    struct termios oldt, newt;
    tcgetattr(STDIN_FILENO, &oldt);
    newt = oldt;
    newt.c_lflag &= ~(ICANON | ECHO);
    tcsetattr(STDIN_FILENO, TCSANOW, &newt);
    fcntl(STDIN_FILENO, F_SETFL, O_NONBLOCK);

    char ch;
    if (read(STDIN_FILENO, &ch, 1) > 0) {
        switch (ch) {
            case 'w': board.snake.changeDirection(Direction::UP); break;
            case 's': board.snake.changeDirection(Direction::DOWN); break;
            case 'a': board.snake.changeDirection(Direction::LEFT); break;
            case 'd': board.snake.changeDirection(Direction::RIGHT); break;
            case 'p': isPaused = !isPaused; break;
            case 'q': isRunning = false; break;
        }
    }
    tcsetattr(STDIN_FILENO, TCSANOW, &oldt);
}

void GameEngine::update() {
    Position head = board.snake.getHead();
    board.snake.move();
    if (board.snake.getHead() == board.food.getLocation()) {
        board.snake.grow();
        board.food.generate(board.getOccupied(), 20, 20);
        score += 10;
    }
}

void GameEngine::checkCollisions() {
    Position head = board.snake.getHead();
    if (!board.isInside(head) || board.snake.isColliding(head) || board.obstacle.isObstacle(head)) {
        isRunning = false;
    }
}

void GameEngine::loadHighScore() {
    std::ifstream in("highscore.txt");
    if (in >> highScore) in.close();
    else highScore = 0;
}

void GameEngine::saveHighScore() {
    if (score > highScore) {
        std::ofstream out("highscore.txt");
        out << score;
        out.close();
    }
}
