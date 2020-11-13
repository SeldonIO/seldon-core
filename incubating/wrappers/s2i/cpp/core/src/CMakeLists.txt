
include(FindProtobuf)
find_package(Protobuf REQUIRED)

find_package(PythonLibs REQUIRED)
#find_package(Python COMPONENTS Interpreter Development REQUIRED)

add_subdirectory(../../pybind11 ${CMAKE_CURRENT_BINARY_DIR}/pybind_build})

file(GLOB SELDON_SRC
    "${CMAKE_CURRENT_SOURCE_DIR}/*.cpp")

file(GLOB SELDON_PROTO_SRC
    "${CMAKE_CURRENT_SOURCE_DIR}/proto/*.cc")


pybind11_add_module(
    seldon
    SHARED ${SELDON_SRC} ${SELDON_PROTO_SRC})

## Build static library as well
#add_library(
#    seldon::seldon ALIAS seldon SHARED
#    ${SELDON_SRC} ${SELDON_PROTO_SRC})

target_include_directories(
    seldon PUBLIC
    $<INSTALL_INTERFACE:include>
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
)

target_link_libraries(
    seldon PRIVATE
    ${PROTOBUF_LIBRARY})

#include_directories(${PYTHON_INCLUDE_DIRS})

#target_link_libraries(
#    seldon PRIVATE 
#    pybind11::module)
#
#pybind11_extension(seldon)
#
#target_link_libraries(
#    seldon PRIVATE 
#    pybind11::lto)
#
#pybind11_strip(seldon)
#
#if(MSVC)
#    target_link_libraries(
#        seldon PRIVATE 
#        pybind11::windows_extras)
#endif()

#set_target_properties(seldon PROPERTIES CXX_VISIBILITY_PRESET "hidden"
#                                        CUDA_VISIBILITY_PRESET "hidden")

add_library(seldon::seldon ALIAS seldon)


