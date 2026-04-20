#pragma once

/**
 * @file DescriptorAllocator.h
 * @brief Alocador de descriptor pools e sets
 */

#include <vulkan/vulkan.h>
#include <vector>
#include <unordered_map>

namespace vklib {

/**
 * @brief Gerencia alocação de VkDescriptorPool e VkDescriptorSet
 *
 * Facilita a criação de descriptor sets reutilizáveis.
 */
class DescriptorAllocator {
public:
    DescriptorAllocator() = default;
    ~DescriptorAllocator();

    // Não permitir cópia
    DescriptorAllocator(const DescriptorAllocator&) = delete;
    DescriptorAllocator& operator=(const DescriptorAllocator&) = delete;

    // ────────────────────────────────────────────────────────────────────────
    // Inicialização
    // ────────────────────────────────────────────────────────────────────────

    /**
     * @brief Inicializa o alocador
     * @param device Dispositivo Vulkan
     * @param maxSets Número máximo de descriptor sets
     */
    void Initialize(VkDevice device, uint32_t maxSets);

    /**
     * @brief Aloca um novo descriptor set
     * @param layout Layout do descriptor set
     * @return VkDescriptorSet alocado
     */
    VkDescriptorSet Allocate(VkDescriptorSetLayout layout);

    /**
     * @brief Libera todos os descriptor sets
     */
    void Reset();

    // ────────────────────────────────────────────────────────────────────────
    // Getters
    // ────────────────────────────────────────────────────────────────────────

    VkDescriptorPool GetPool() const { return m_pool; }

private:
    VkDevice m_device = VK_NULL_HANDLE;
    VkDescriptorPool m_pool = VK_NULL_HANDLE;
};

}  // namespace vklib
