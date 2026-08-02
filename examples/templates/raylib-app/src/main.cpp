#include "raylib.h"

int main() {
    constexpr int screenWidth = 800;
    constexpr int screenHeight = 450;

    InitWindow(screenWidth, screenHeight, "{{.Name}}");
    SetTargetFPS(60);

    while (!WindowShouldClose()) {
        BeginDrawing();
        ClearBackground(RAYWHITE);
        DrawText("{{.Name}}", 20, 20, 20, DARKGRAY);
        DrawText("Gerado pelo cpp-gen — pressione ESC para sair", 20, 50, 10, GRAY);
        DrawCircle(screenWidth / 2, screenHeight / 2, 50, MAROON);
        EndDrawing();
    }

    CloseWindow();
    return 0;
}
