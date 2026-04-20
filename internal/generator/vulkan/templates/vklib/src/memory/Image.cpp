#include <vklib/common/Defines.h>
#include <vklib/memory/Image.h>

#include <cassert>

namespace vklib {

/**
 * @brief Converte ImageFormat para VkFormat
 */
static VkFormat GetVkFormat(ImageFormat format)
{
    switch (format) {
        case ImageFormat::RGBA8:
            return VK_FORMAT_R8G8B8A8_SRGB;
        case ImageFormat::Depth32F:
            return VK_FORMAT_D32_SFLOAT;
        default:
            return VK_FORMAT_UNDEFINED;
    }
}

/**
 * @brief Retorna as flags de uso baseado no tipo de imagem
 */
static VkImageUsageFlags GetImageUsageFlags(ImageFormat format)
{
    switch (format) {
        case ImageFormat::RGBA8:
            // Textura: pode ser alvo de transfer e sampled em shaders
            return VK_IMAGE_USAGE_TRANSFER_DST_BIT | VK_IMAGE_USAGE_SAMPLED_BIT;

        case ImageFormat::Depth32F:
            // Depth buffer: alvo de renderização
            return VK_IMAGE_USAGE_DEPTH_STENCIL_ATTACHMENT_BIT;

        default:
            return 0;
    }
}

/**
 * @brief Retorna os aspect flags para criar image views
 */
static VkImageAspectFlags GetImageAspect(ImageFormat format)
{
    switch (format) {
        case ImageFormat::RGBA8:
            return VK_IMAGE_ASPECT_COLOR_BIT;

        case ImageFormat::Depth32F:
            return VK_IMAGE_ASPECT_DEPTH_BIT;

        default:
            return 0;
    }
}

Image::~Image()
{
    if (m_imageView != VK_NULL_HANDLE && m_allocator != VK_NULL_HANDLE) {
        // Precisamos do device para destruir image view, mas não temos
        // Por isso é importante chamar Destroy() ou deixar o move semantics funcionar
        VK_LOG_WARN("Image view não foi destruída - use move semantics ou destrua manualmente");
    }
}

Image::Image(Image&& other) noexcept
    : m_image(std::exchange(other.m_image, VK_NULL_HANDLE)),
      m_imageView(std::exchange(other.m_imageView, VK_NULL_HANDLE)),
      m_allocation(std::exchange(other.m_allocation, VK_NULL_HANDLE)),
      m_allocator(std::exchange(other.m_allocator, VK_NULL_HANDLE)),
      m_width(std::exchange(other.m_width, 0)), m_height(std::exchange(other.m_height, 0)),
      m_format(std::exchange(other.m_format, ImageFormat::RGBA8)),
      m_currentLayout(std::exchange(other.m_currentLayout, VK_IMAGE_LAYOUT_UNDEFINED))
{}

Image& Image::operator=(Image&& other) noexcept
{
    if (this != &other) {
        // Não destruímos aqui porque não temos o device
        // O destrutor do outro cuida disso

        m_image = std::exchange(other.m_image, VK_NULL_HANDLE);
        m_imageView = std::exchange(other.m_imageView, VK_NULL_HANDLE);
        m_allocation = std::exchange(other.m_allocation, VK_NULL_HANDLE);
        m_allocator = std::exchange(other.m_allocator, VK_NULL_HANDLE);
        m_width = std::exchange(other.m_width, 0);
        m_height = std::exchange(other.m_height, 0);
        m_format = std::exchange(other.m_format, ImageFormat::RGBA8);
        m_currentLayout = std::exchange(other.m_currentLayout, VK_IMAGE_LAYOUT_UNDEFINED);
    }
    return *this;
}

void Image::Create(VkDevice device,
                   VmaAllocator allocator,
                   uint32_t width,
                   uint32_t height,
                   ImageFormat format,
                   VkImageUsageFlags usage)
{
    assert(device != VK_NULL_HANDLE && "Device deve ser válido");
    assert(allocator != VK_NULL_HANDLE && "VmaAllocator deve ser válido");
    assert(width > 0 && height > 0 && "Dimensões devem ser > 0");

    m_allocator = allocator;
    m_width = width;
    m_height = height;
    m_format = format;
    m_currentLayout = VK_IMAGE_LAYOUT_UNDEFINED;

    VkFormat vkFormat = GetVkFormat(format);
    assert(vkFormat != VK_FORMAT_UNDEFINED && "Formato de imagem inválido");

    VkImageCreateInfo imageInfo{};
    imageInfo.sType = VK_STRUCTURE_TYPE_IMAGE_CREATE_INFO;
    imageInfo.imageType = VK_IMAGE_TYPE_2D;
    imageInfo.extent.width = width;
    imageInfo.extent.height = height;
    imageInfo.extent.depth = 1;
    imageInfo.mipLevels = 1;
    imageInfo.arrayLayers = 1;
    imageInfo.format = vkFormat;
    imageInfo.tiling = VK_IMAGE_TILING_OPTIMAL;
    imageInfo.initialLayout = VK_IMAGE_LAYOUT_UNDEFINED;
    imageInfo.usage = usage | GetImageUsageFlags(format);
    imageInfo.samples = VK_SAMPLE_COUNT_1_BIT;
    imageInfo.sharingMode = VK_SHARING_MODE_EXCLUSIVE;

    VmaAllocationCreateInfo allocInfo{};
    allocInfo.usage = VMA_MEMORY_USAGE_AUTO;
    allocInfo.requiredFlags = VK_MEMORY_PROPERTY_DEVICE_LOCAL_BIT;

    VkResult result =
        vmaCreateImage(allocator, &imageInfo, &allocInfo, &m_image, &m_allocation, nullptr);

    if (result != VK_SUCCESS) {
        VK_LOG_ERROR("Falha ao alocar imagem: " << static_cast<int>(result));
        m_image = VK_NULL_HANDLE;
        m_allocation = VK_NULL_HANDLE;
        return;
    }

    VkImageViewCreateInfo viewInfo{};
    viewInfo.sType = VK_STRUCTURE_TYPE_IMAGE_VIEW_CREATE_INFO;
    viewInfo.image = m_image;
    viewInfo.viewType = VK_IMAGE_VIEW_TYPE_2D;
    viewInfo.format = vkFormat;
    viewInfo.components.r = VK_COMPONENT_SWIZZLE_IDENTITY;
    viewInfo.components.g = VK_COMPONENT_SWIZZLE_IDENTITY;
    viewInfo.components.b = VK_COMPONENT_SWIZZLE_IDENTITY;
    viewInfo.components.a = VK_COMPONENT_SWIZZLE_IDENTITY;
    viewInfo.subresourceRange.aspectMask = GetImageAspect(format);
    viewInfo.subresourceRange.baseMipLevel = 0;
    viewInfo.subresourceRange.levelCount = 1;
    viewInfo.subresourceRange.baseArrayLayer = 0;
    viewInfo.subresourceRange.layerCount = 1;

    VK_CHECK(vkCreateImageView(device, &viewInfo, nullptr, &m_imageView));

    const char* formatStr = "";
    switch (format) {
        case ImageFormat::RGBA8:
            formatStr = "RGBA8";
            break;
        case ImageFormat::Depth32F:
            formatStr = "Depth32F";
            break;
    }

    VK_LOG_INFO("Imagem " << formatStr << " alocada: " << width << "x" << height);
}

void Image::TransitionLayout(VkCommandBuffer cmd, VkImageLayout oldLayout, VkImageLayout newLayout)
{
    assert(cmd != VK_NULL_HANDLE && "Command buffer deve ser válido");
    assert(m_image != VK_NULL_HANDLE && "Imagem não foi criada");

    VkPipelineStageFlags srcStage = 0;
    VkPipelineStageFlags dstStage = 0;
    VkAccessFlags srcAccess = 0;
    VkAccessFlags dstAccess = 0;

    // Transição FROM
    if (oldLayout == VK_IMAGE_LAYOUT_UNDEFINED) {
        srcStage = VK_PIPELINE_STAGE_TOP_OF_PIPE_BIT;
        srcAccess = 0;
    }
    else if (oldLayout == VK_IMAGE_LAYOUT_TRANSFER_DST_OPTIMAL) {
        srcStage = VK_PIPELINE_STAGE_TRANSFER_BIT;
        srcAccess = VK_ACCESS_TRANSFER_WRITE_BIT;
    }
    else if (oldLayout == VK_IMAGE_LAYOUT_SHADER_READ_ONLY_OPTIMAL) {
        srcStage = VK_PIPELINE_STAGE_FRAGMENT_SHADER_BIT;
        srcAccess = VK_ACCESS_SHADER_READ_BIT;
    }
    else if (oldLayout == VK_IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL) {
        srcStage = VK_PIPELINE_STAGE_COLOR_ATTACHMENT_OUTPUT_BIT;
        srcAccess = VK_ACCESS_COLOR_ATTACHMENT_WRITE_BIT;
    }
    else if (oldLayout == VK_IMAGE_LAYOUT_DEPTH_STENCIL_ATTACHMENT_OPTIMAL) {
        srcStage = VK_PIPELINE_STAGE_EARLY_FRAGMENT_TESTS_BIT;
        srcAccess = VK_ACCESS_DEPTH_STENCIL_ATTACHMENT_WRITE_BIT;
    }
    else {
        srcStage = VK_PIPELINE_STAGE_BOTTOM_OF_PIPE_BIT;
        srcAccess = 0;
    }

    if (newLayout == VK_IMAGE_LAYOUT_TRANSFER_DST_OPTIMAL) {
        dstStage = VK_PIPELINE_STAGE_TRANSFER_BIT;
        dstAccess = VK_ACCESS_TRANSFER_WRITE_BIT;
    }
    else if (newLayout == VK_IMAGE_LAYOUT_SHADER_READ_ONLY_OPTIMAL) {
        dstStage = VK_PIPELINE_STAGE_FRAGMENT_SHADER_BIT;
        dstAccess = VK_ACCESS_SHADER_READ_BIT;
    }
    else if (newLayout == VK_IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL) {
        dstStage = VK_PIPELINE_STAGE_COLOR_ATTACHMENT_OUTPUT_BIT;
        dstAccess = VK_ACCESS_COLOR_ATTACHMENT_WRITE_BIT;
    }
    else if (newLayout == VK_IMAGE_LAYOUT_DEPTH_STENCIL_ATTACHMENT_OPTIMAL) {
        dstStage = VK_PIPELINE_STAGE_EARLY_FRAGMENT_TESTS_BIT;
        dstAccess = VK_ACCESS_DEPTH_STENCIL_ATTACHMENT_WRITE_BIT |
                    VK_ACCESS_DEPTH_STENCIL_ATTACHMENT_READ_BIT;
    }
    else if (newLayout == VK_IMAGE_LAYOUT_PRESENT_SRC_KHR) {
        dstStage = VK_PIPELINE_STAGE_BOTTOM_OF_PIPE_BIT;
        dstAccess = 0;
    }
    else {
        dstStage = VK_PIPELINE_STAGE_BOTTOM_OF_PIPE_BIT;
        dstAccess = 0;
    }

    VkImageMemoryBarrier barrier{};
    barrier.sType = VK_STRUCTURE_TYPE_IMAGE_MEMORY_BARRIER;
    barrier.oldLayout = oldLayout;
    barrier.newLayout = newLayout;
    barrier.srcQueueFamilyIndex = VK_QUEUE_FAMILY_IGNORED;
    barrier.dstQueueFamilyIndex = VK_QUEUE_FAMILY_IGNORED;
    barrier.image = m_image;
    barrier.subresourceRange.aspectMask = GetImageAspect(m_format);
    barrier.subresourceRange.baseMipLevel = 0;
    barrier.subresourceRange.levelCount = 1;
    barrier.subresourceRange.baseArrayLayer = 0;
    barrier.subresourceRange.layerCount = 1;
    barrier.srcAccessMask = srcAccess;
    barrier.dstAccessMask = dstAccess;

    vkCmdPipelineBarrier(cmd,
                         srcStage,
                         dstStage,
                         0,  // dependencyFlags
                         0,
                         nullptr,  // memory barriers
                         0,
                         nullptr,  // buffer memory barriers
                         1,
                         &barrier  // image memory barriers
    );

    // Atualizar layout atual
    m_currentLayout = newLayout;
}
}  // namespace vklib
