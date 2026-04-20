#pragma once

/**
 * @file SystemInfo.hpp
 * @brief Utilitário para obter informações do sistema Vulkan
 */

#include <vulkan/vulkan.h>

#include <memory>
#include <string>
#include <vector>

namespace vklib {

/**
 * @brief Informações do sistema Vulkan
 *
 * Fornece acesso a layers, extensões e capacidades do sistema
 */
class SystemInfo
{
public:
    /**
     * @brief Obtém informações do sistema
     * @return true se bem-sucedido
     */
    static bool Initialize();

    /**
     * @brief Verifica se uma layer está disponível
     */
    static bool IsLayerAvailable(const std::string& layerName);

    /**
     * @brief Verifica se uma extensão está disponível
     */
    static bool IsExtensionAvailable(const std::string& extensionName);

    /**
     * @brief Lista todas as layers disponíveis
     */
    static const std::vector<std::string>& GetAvailableLayers();

    /**
     * @brief Lista todas as extensões disponíveis
     */
    static const std::vector<std::string>& GetAvailableExtensions();

    /**
     * @brief Verifica se validation layers estão disponíveis
     */
    static bool HasValidationLayers();

    /**
     * @brief Imprime todas as informações do sistema
     */
    static void PrintSystemInfo();

private:
    static std::vector<std::string> s_availableLayers;
    static std::vector<std::string> s_availableExtensions;
    static bool s_validationLayersAvailable;
    static bool s_initialized;
};

}  // namespace vklib
