cmake_minimum_required(VERSION 3.10)
project(SnakeGame)

set(CMAKE_CXX_STANDARD 17)

add_executable(SnakeGame
    main.cpp
    Snake.cpp
    Food.cpp
    Obstacle.cpp
    GameBoard.cpp
    GameEngine.cpp
)
