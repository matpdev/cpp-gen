#pragma once

/**
 * @file Defines.h
 * @brief Definições e macros globais da biblioteca VulkanLib
 */

// ═════════════════════════════════════════════════════════════════════════════
// Macros de Validação Vulkan
// ═════════════════════════════════════════════════════════════════════════════

#include <vulkan/vulkan.h>
#include <iostream>
#include <cstdlib>

#define VK_CHECK(x)                                                          \
    do {                                                                     \
        VkResult err = x;                                                    \
        if (err != VK_SUCCESS) {                                             \
            std::cerr << "❌ Erro Vulkan: código " << static_cast<int>(err) \
                      << " em " << __FILE__ << ":" << __LINE__ << "\n";    \
            std::abort();                                                    \
        }                                                                    \
    } while (0)

#define VK_LOG_INFO(msg)    std::cout << "ℹ️  [INFO] " << msg << "\n"
#define VK_LOG_WARN(msg)    std::cout << "⚠️  [WARN] " << msg << "\n"
#define VK_LOG_ERROR(msg)   std::cerr << "❌ [ERROR] " << msg << "\n"

// ═════════════════════════════════════════════════════════════════════════════
// Constantes
// ═════════════════════════════════════════════════════════════════════════════

constexpr int FRAMES_IN_FLIGHT = 2;  // Double buffering

namespace vklib {
    // Versão da biblioteca
    constexpr int VERSION_MAJOR = 0;
    constexpr int VERSION_MINOR = 1;
    constexpr int VERSION_PATCH = 0;
}
