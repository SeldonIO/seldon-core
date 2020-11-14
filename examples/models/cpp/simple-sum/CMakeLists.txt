cmake_minimum_required(VERSION 3.4.1)
project(seldon_custom_model VERSION 0.0.1)

set(CMAKE_CXX_STANDARD 14)

add_subdirectory(
    _cpp/
    ${CMAKE_CURRENT_BINARY_DIR}/seldon_build})

pybind11_add_module(
    SeldonPackage
    SeldonPackage.cpp)

target_link_libraries(
    SeldonPackage PRIVATE
    seldon::seldon)


