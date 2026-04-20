#include "vklib/common/Defines.h"
#include "vklib/system/SystemInfo.h"

#include <algorithm>
#include <iostream>

#include <VkBootstrap.h>

namespace vklib {

// Static member initialization
std::vector<std::string> SystemInfo::s_availableLayers;
std::vector<std::string> SystemInfo::s_availableExtensions;
bool SystemInfo::s_validationLayersAvailable = false;
bool SystemInfo::s_initialized = false;

// ═════════════════════════════════════════════════════════════════════════════
// Inicialização
// ═════════════════════════════════════════════════════════════════════════════

bool SystemInfo::Initialize()
{
    if (s_initialized) {
        return true;
    }

    s_availableLayers.clear();
    s_availableExtensions.clear();

    // ─────────────────────────────────────────────────────────────────────────
    // Obter layers
    // ─────────────────────────────────────────────────────────────────────────

    uint32_t layerCount = 0;
    vkEnumerateInstanceLayerProperties(&layerCount, nullptr);

    std::vector<VkLayerProperties> layers(layerCount);
    vkEnumerateInstanceLayerProperties(&layerCount, layers.data());

    // VK_LOG_INFO("Layers disponíveis: " << layerCount);
    // for (const auto& layer : layers) {
    //     s_availableLayers.emplace_back(layer.layerName);
    //     VK_LOG_INFO("  → " << layer.layerName << " (v" << layer.specVersion << ")");
    // }

    // Verificar se validation layers estão disponíveis
    VK_LOG_INFO("ESTARTED");
    s_validationLayersAvailable = IsLayerAvailable("VK_LAYER_KHRONOS_validation");

    // ─────────────────────────────────────────────────────────────────────────
    // Obter extensões
    // ─────────────────────────────────────────────────────────────────────────

    uint32_t extensionCount = 0;
    vkEnumerateInstanceExtensionProperties(nullptr, &extensionCount, nullptr);

    std::vector<VkExtensionProperties> extensions(extensionCount);
    vkEnumerateInstanceExtensionProperties(nullptr, &extensionCount, extensions.data());

    VK_LOG_INFO("Extensões disponíveis: " << extensionCount);
    for (const auto& ext : extensions) {
        s_availableExtensions.emplace_back(ext.extensionName);
        VK_LOG_INFO("  → " << ext.extensionName);
    }

    s_initialized = true;
    return true;
}

// ═════════════════════════════════════════════════════════════════════════════
// Verificações
// ═════════════════════════════════════════════════════════════════════════════

bool SystemInfo::IsLayerAvailable(const std::string& layerName)
{
    // if (!s_initialized) {
    //     Initialize();
    // }

    return std::find(s_availableLayers.begin(), s_availableLayers.end(), layerName) !=
           s_availableLayers.end();
}

bool SystemInfo::IsExtensionAvailable(const std::string& extensionName)
{
    if (!s_initialized) {
        Initialize();
    }

    return std::find(s_availableExtensions.begin(), s_availableExtensions.end(), extensionName) !=
           s_availableExtensions.end();
}

// ═════════════════════════════════════════════════════════════════════════════
// Getters
// ═════════════════════════════════════════════════════════════════════════════

const std::vector<std::string>& SystemInfo::GetAvailableLayers()
{
    if (!s_initialized) {
        Initialize();
    }
    return s_availableLayers;
}

const std::vector<std::string>& SystemInfo::GetAvailableExtensions()
{
    if (!s_initialized) {
        Initialize();
    }
    return s_availableExtensions;
}

bool SystemInfo::HasValidationLayers()
{
    if (!s_initialized) {
        Initialize();
    }
    return s_validationLayersAvailable;
}

// ═════════════════════════════════════════════════════════════════════════════
// Print
// ═════════════════════════════════════════════════════════════════════════════

void SystemInfo::PrintSystemInfo()
{
    if (!s_initialized) {
        Initialize();
    }

    std::cout << "\n";
    std::cout << "═════════════════════════════════════════════════════════════\n";
    std::cout << "                   VULKAN SYSTEM INFO\n";
    std::cout << "═════════════════════════════════════════════════════════════\n";

    // Layers
    std::cout << "\n▶ LAYERS (" << s_availableLayers.size() << ")\n";
    for (const auto& layer : s_availableLayers) {
        std::cout << "  ✓ " << layer << "\n";
    }

    if (s_validationLayersAvailable) {
        std::cout << "\n  ✓ Validation Layers: DISPONÍVEL\n";
    }
    else {
        std::cout << "\n  ✗ Validation Layers: NÃO DISPONÍVEL\n";
    }

    // Extensions
    std::cout << "\n▶ EXTENSIONS (" << s_availableExtensions.size() << ")\n";
    for (const auto& ext : s_availableExtensions) {
        std::cout << "  ✓ " << ext << "\n";
    }

    // Key extensions
    std::cout << "\n▶ KEY EXTENSIONS\n";
    std::cout << "  " << (IsExtensionAvailable(VK_KHR_SURFACE_EXTENSION_NAME) ? "✓" : "✗") << " "
              << VK_KHR_SURFACE_EXTENSION_NAME << "\n";
    std::cout << "  " << (IsExtensionAvailable(VK_EXT_DEBUG_UTILS_EXTENSION_NAME) ? "✓" : "✗")
              << " " << VK_EXT_DEBUG_UTILS_EXTENSION_NAME << "\n";
    std::cout << "  "
              << (IsExtensionAvailable(VK_KHR_GET_PHYSICAL_DEVICE_PROPERTIES_2_EXTENSION_NAME)
                      ? "✓"
                      : "✗")
              << " " << VK_KHR_GET_PHYSICAL_DEVICE_PROPERTIES_2_EXTENSION_NAME << "\n";

    std::cout << "\n═════════════════════════════════════════════════════════════\n\n";
}

}  // namespace vklib
