/**
 * @file main.cpp
 * @brief {{.NamePascal}} — ponto de entrada da aplicação Vulkan
 *
 * Gerado pelo cpp-gen com o template Vulkan.
 * Este arquivo é o ponto de partida da sua aplicação.
 * A biblioteca vklib (em vklib/) já provê a abstração do engine Vulkan.
 */

#include <vklib/VulkanLib.h>

#include <iostream>

// Fallback caso a macro SHADER_DIR não venha do CMake
#ifndef SHADER_DIR
#define SHADER_DIR "./shaders/"
#endif

using namespace vklib;

int main([[maybe_unused]] int argc, [[maybe_unused]] char* argv[])
{
    // ── Inicialização ─────────────────────────────────────────────────────────
    EngineContext engine;

    if (!engine.Initialize("{{.NamePascal}}", 800, 600)) {
        std::cerr << "[{{.NamePascal}}] Falha ao inicializar o engine Vulkan.
";
        return 1;
    }

    std::cout << "[{{.NamePascal}}] Engine inicializado. Iniciando loop...
";

    // ── Loop de renderização ──────────────────────────────────────────────────
    while (engine.IsRunning()) {
        FrameData& frame = engine.BeginFrame();

        if (!engine.IsFrameReady()) {
            engine.EndFrame();
            continue;
        }

        // TODO: despache seus comandos Vulkan aqui usando:
        //   VkCommandBuffer cmd = frame.GetCommandBuffer();
        //   vkCmdBindPipeline(cmd, VK_PIPELINE_BIND_POINT_GRAPHICS, pipeline);
        //   vkCmdDraw(cmd, vertexCount, 1, 0, 0);
        (void)frame;

        engine.EndFrame();
    }

    // ── Cleanup ───────────────────────────────────────────────────────────────
    vkDeviceWaitIdle(engine.GetDevice());
    engine.Shutdown();

    std::cout << "[{{.NamePascal}}] Aplicação finalizada.
";
    return 0;
}
