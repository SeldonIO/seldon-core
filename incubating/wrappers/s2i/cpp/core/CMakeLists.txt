cmake_minimum_required(VERSION 3.4.1)
project(seldon VERSION 0.0.1)

set(CMAKE_CXX_FLAGS_DEBUG "${CMAKE_CXX_FLAGS_DEBUG} -DDEBUG=1 -fPIC")
set(CMAKE_CXX_FLAGS_RELEASE "${CMAKE_CXX_FLAGS_RELEASE} -DRELEASE=1 -fPIC")

set(CMAKE_CXX_STANDARD 14)

add_subdirectory(src)

add_subdirectory(test)

