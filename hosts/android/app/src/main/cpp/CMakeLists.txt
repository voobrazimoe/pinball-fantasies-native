cmake_minimum_required(VERSION 3.22.1)
project(pinball_android LANGUAGES C CXX)

find_package(game-activity REQUIRED CONFIG)

add_library(pinball_android SHARED main.cpp)

target_compile_features(pinball_android PRIVATE cxx_std_17)
target_compile_options(pinball_android PRIVATE -Wall -Wextra -Werror)
target_link_options(pinball_android PRIVATE
    "LINKER:-u,GameActivity_onCreate"
    "LINKER:-u,Java_com_google_androidgamesdk_GameActivity_initializeNativeCode"
    "LINKER:-z,max-page-size=16384"
    "LINKER:-z,common-page-size=16384")

target_link_libraries(pinball_android PRIVATE
    android
    c++
    EGL
    GLESv2
    game-activity::game-activity_static
    log)
