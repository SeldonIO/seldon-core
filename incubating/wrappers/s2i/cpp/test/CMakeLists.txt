
find_package(Protobuf REQUIRED)
find_package(PythonLibs REQUIRED)

add_executable(seldon-test
    Main.cpp catch_amalgamated.cpp)

target_include_directories(
    seldon-test PUBLIC
    $<INSTALL_INTERFACE:include>
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>)

target_include_directories(
    seldon-test PUBLIC
    ${PROJECT_SOURCE_DIR}/src/include)

include_directories(${PYTHON_INCLUDE_DIRS})

target_link_libraries(
    seldon-test
    ${PROTOBUF_LIBRARIES}
    ${PYTHON_LIBRARIES}
    seldon
)




