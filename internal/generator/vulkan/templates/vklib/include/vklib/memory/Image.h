#pragma once

/**
 * @file Image.h
 * @brief Abstração para imagens (texturas, depth buffers) via VMA
 */

#include "vklib/common/Types.h"

#include <vulkan/vulkan.h>

#include <cstdint>
#include <utility>

#include <vk_mem_alloc.h>

namespace vklib {

/**
 * @brief Imagem GPU gerenciada por VMA
 *
 * Encapsula VkImage, VkImageView e VmaAllocation.
 * Suporta texturas 2D e depth buffers.
 */
class Image
{
public:
    Image() = default;
    ~Image();

    // Não permitir cópia
    Image(const Image&) = delete;
    Image& operator=(const Image&) = delete;

    // Mover é permitido
    Image(Image&& other) noexcept;
    Image& operator=(Image&& other) noexcept;

    // ────────────────────────────────────────────────────────────────────────
    // Criação
    // ────────────────────────────────────────────────────────────────────────

    /**
     * @brief Cria uma imagem na GPU
     * @param allocator Alocador VMA
     * @param width Largura em pixels
     * @param height Altura em pixels
     * @param format Formato de imagem
     * @param usage Flags de uso (COLOR_ATTACHMENT, TRANSFER_DST, etc.)
     */
    void Create(VkDevice device,
                VmaAllocator allocator,
                uint32_t width,
                uint32_t height,
                ImageFormat format,
                VkImageUsageFlags usage);

    /**
     * @brief Faz transição de layout da imagem (para renderização ou shader sampling)
     * @param cmd Command buffer onde gravar o comando
     * @param oldLayout Layout anterior
     * @param newLayout Layout desejado
     */
    void TransitionLayout(VkCommandBuffer cmd, VkImageLayout oldLayout, VkImageLayout newLayout);

    // ────────────────────────────────────────────────────────────────────────
    // Getters (Escape Hatches)
    // ────────────────────────────────────────────────────────────────────────

    VkImage GetImage() const
    {
        return m_image;
    }
    VkImageView GetImageView() const
    {
        return m_imageView;
    }
    uint32_t GetWidth() const
    {
        return m_width;
    }
    uint32_t GetHeight() const
    {
        return m_height;
    }
    ImageFormat GetFormat() const
    {
        return m_format;
    }
    VkImageLayout GetCurrentLayout() const
    {
        return m_currentLayout;
    }

private:
    VkImage m_image = VK_NULL_HANDLE;
    VkImageView m_imageView = VK_NULL_HANDLE;
    VmaAllocation m_allocation = VK_NULL_HANDLE;
    VmaAllocator m_allocator = VK_NULL_HANDLE;
    uint32_t m_width = 0;
    uint32_t m_height = 0;
    ImageFormat m_format = ImageFormat::RGBA8;
    VkImageLayout m_currentLayout = VK_IMAGE_LAYOUT_UNDEFINED;
};

}  // namespace vklib
