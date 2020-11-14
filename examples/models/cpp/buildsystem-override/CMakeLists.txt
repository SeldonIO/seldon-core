cmake_minimum_required(VERSION 3.4.1)
project(seldon_custom_model VERSION 0.0.1)

set(CMAKE_CXX_STANDARD 14)

find_package(seldon REQUIRED)
find_package(pybind11 REQUIRED)

pybind11_add_module(
    CustomSeldonPackage
    Main.cpp)

target_link_libraries(
    CustomSeldonPackage PRIVATE
    seldon::seldon)


