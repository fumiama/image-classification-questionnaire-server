cmake_minimum_required(VERSION 3.0.0)
project(base16384 VERSION 2.0)

add_library(base14 MODULE ../.base14/base16384.c)

IF(CMAKE_SIZEOF_VOID_P EQUAL 8)
    add_definitions("-DCPUBIT64")
    add_subdirectory("./.base14/64")
    target_link_libraries(base14 base1464)
ELSE()
    add_definitions("-DCPUBIT32")
    add_subdirectory("./.base14/32")
    target_link_libraries(base14 base1432)
ENDIF()
