#pragma once

/**
 * @file VkCheck.h
 * @brief Funções auxiliares para verificação de erros Vulkan
 */

#include <vulkan/vulkan.h>
#include <string_view>

namespace vklib {

/**
 * @brief Converte um VkResult para uma string legível
 * @param result O código de erro Vulkan
 * @return String descrevendo o erro
 */
std::string_view VkResultToString(VkResult result);

/**
 * @brief Loga um erro Vulkan com contexto detalhado
 * @param result Código do erro
 * @param file Arquivo onde ocorreu
 * @param line Linha onde ocorreu
 * @param context Descrição do contexto (ex: "vkCreateDevice")
 */
void LogVulkanError(VkResult result, const char* file, int line,
                    std::string_view context);

}  // namespace vklib
