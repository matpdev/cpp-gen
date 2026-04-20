#pragma once

/**
 * @file Buffer.h
 * @brief Abstração para buffers GPU via Vulkan Memory Allocator (VMA)
 */

#include "vklib/common/Types.h"

#include <vulkan/vulkan.h>

#include <cstring>
#include <memory>
#include <utility>

#include <vk_mem_alloc.h>

namespace vklib {

/**
 * @brief Buffer GPU gerenciado por VMA
 *
 * Encapsula VkBuffer + VmaAllocation para gerenciar memória GPU.
 * Suporta Vertex, Index, Uniform e Staging buffers.
 */
class Buffer
{
public:
    Buffer() = default;
    ~Buffer();

    // Não permitir cópia
    Buffer(const Buffer&) = delete;
    Buffer& operator=(const Buffer&) = delete;

    // Mover é permitido
    Buffer(Buffer&& other) noexcept;
    Buffer& operator=(Buffer&& other) noexcept;

    // ────────────────────────────────────────────────────────────────────────
    // Criação
    // ────────────────────────────────────────────────────────────────────────

    /**
     * @brief Cria um buffer na GPU
     * @param allocator Alocador VMA
     * @param size Tamanho em bytes
     * @param type Tipo de buffer (Vertex, Index, Uniform, Staging)
     */
    void Create(VmaAllocator allocator, VkDeviceSize size, BufferType type);

    /**
     * @brief Copia dados para o buffer (só funciona com Staging buffer)
     * @param data Ponteiro para dados
     * @param size Tamanho dos dados
     */
    void SetData(const void* data, VkDeviceSize size);

    /**
     * @brief Mapeia o buffer para acesso da CPU (se for staging ou host-visible)
     * @return Ponteiro para a memória mapeada
     */
    void* Map();

    /**
     * @brief Desmapeia o buffer após edição da CPU
     */
    void Unmap();

    // ────────────────────────────────────────────────────────────────────────
    // Getters (Escape Hatches)
    // ────────────────────────────────────────────────────────────────────────

    VkBuffer GetBuffer() const
    {
        return m_buffer;
    }
    VkDeviceSize GetSize() const
    {
        return m_size;
    }
    BufferType GetType() const
    {
        return m_type;
    }

private:
    VkBuffer m_buffer = VK_NULL_HANDLE;
    VmaAllocation m_allocation = VK_NULL_HANDLE;
    VmaAllocator m_allocator = VK_NULL_HANDLE;
    VkDeviceSize m_size = 0;
    BufferType m_type = BufferType::Vertex;
};

}  // namespace vklib
