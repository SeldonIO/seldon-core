cmake_minimum_required(VERSION 3.4.1)
project(seldon_custom_model VERSION 0.0.1)

set(CMAKE_CXX_STANDARD 14)

add_subdirectory(
    ../../../../incubating/wrappers/s2i/cpp/core/
    ${CMAKE_CURRENT_BINARY_DIR}/seldon_build})

#get_target_property(seldon_INCLUDE_DIRS seldon INCLUDE_DIRECTORIES)
#message(STATUS "TESTING THIS")
#message(STATUS ${seldon_INCLUDE_DIRS})

pybind11_add_module(
    seldon_custom_model
    CustomModel.cpp)

target_link_libraries(
    seldon_custom_model PRIVATE
    seldon::seldon)


