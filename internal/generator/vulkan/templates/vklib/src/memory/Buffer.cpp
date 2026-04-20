#include "vklib/common/Types.h"

#include <vklib/common/Defines.h>
#include <vklib/memory/Buffer.h>
#include <vulkan/vulkan_core.h>

namespace vklib {
Buffer::~Buffer()
{
    if (m_buffer != VK_NULL_HANDLE && m_allocator != VK_NULL_HANDLE) {
        vmaDestroyBuffer(m_allocator, m_buffer, m_allocation);
        m_buffer = VK_NULL_HANDLE;
        m_allocation = VK_NULL_HANDLE;
    }
}

Buffer::Buffer(Buffer&& other) noexcept
    : m_buffer(std::exchange(other.m_buffer, VK_NULL_HANDLE)),
      m_allocation(std::exchange(other.m_allocation, VK_NULL_HANDLE)),
      m_allocator(std::exchange(other.m_allocator, VK_NULL_HANDLE)),
      m_size(std::exchange(other.m_size, 0)),
      m_type(std::exchange(other.m_type, BufferType::Vertex))
{}

Buffer& Buffer::operator=(Buffer&& other) noexcept
{
    if (this != &other) {
        // Destruir recursos existentes
        if (m_buffer != VK_NULL_HANDLE && m_allocator != VK_NULL_HANDLE) {
            vmaDestroyBuffer(m_allocator, m_buffer, m_allocation);
        }

        // Mover recursos
        m_buffer = std::exchange(other.m_buffer, VK_NULL_HANDLE);
        m_allocation = std::exchange(other.m_allocation, VK_NULL_HANDLE);
        m_allocator = std::exchange(other.m_allocator, VK_NULL_HANDLE);
        m_size = std::exchange(other.m_size, 0);
        m_type = std::exchange(other.m_type, BufferType::Vertex);
    }
    return *this;
}

static VkBufferUsageFlags GetBufferUsageFlags(BufferType type)
{
    switch (type) {
        case BufferType::Vertex:
            return VK_BUFFER_USAGE_VERTEX_BUFFER_BIT | VK_BUFFER_USAGE_TRANSFER_DST_BIT;

        case BufferType::Index:
            return VK_BUFFER_USAGE_INDEX_BUFFER_BIT | VK_BUFFER_USAGE_TRANSFER_DST_BIT;

        case BufferType::Uniform:
            return VK_BUFFER_USAGE_UNIFORM_BUFFER_BIT | VK_BUFFER_USAGE_TRANSFER_DST_BIT;

        case BufferType::Staging:
            return VK_BUFFER_USAGE_TRANSFER_SRC_BIT;

        default:
            return 0;
    }
}

static VmaAllocationCreateFlags GetAllocationFlags(BufferType type)
{
    switch (type) {
        case BufferType::Staging:
            // Staging buffer precisa estar na memória host-visible e host-coherent
            return VMA_ALLOCATION_CREATE_HOST_ACCESS_SEQUENTIAL_WRITE_BIT |
                   VMA_ALLOCATION_CREATE_MAPPED_BIT;

        case BufferType::Uniform:
            // Uniform buffer também pode ser host-visible para CPU updates
            return VMA_ALLOCATION_CREATE_HOST_ACCESS_RANDOM_BIT | VMA_ALLOCATION_CREATE_MAPPED_BIT;

        case BufferType::Vertex:
        case BufferType::Index:
        default:
            // GPU-only, sem acesso CPU direto (usa staging buffer para transfers)
            return 0;
    }
}

static VkMemoryPropertyFlags GetMemoryPropertyFlags(BufferType type)
{
    switch (type) {
        case BufferType::Staging:
            return VK_MEMORY_PROPERTY_HOST_VISIBLE_BIT | VK_MEMORY_PROPERTY_HOST_COHERENT_BIT;

        case BufferType::Uniform:
            return VK_MEMORY_PROPERTY_HOST_VISIBLE_BIT | VK_MEMORY_PROPERTY_HOST_COHERENT_BIT;

        case BufferType::Vertex:
        case BufferType::Index:
        default:
            return VK_MEMORY_PROPERTY_DEVICE_LOCAL_BIT;
    }
}

void Buffer::Create(VmaAllocator allocator, VkDeviceSize size, BufferType type)
{
    assert(allocator != VK_NULL_HANDLE && "VmaAllocator deve ser válido");
    assert(size > 0 && "Tamanho do buffer deve ser > 0");

    // Se já há um buffer alocado, destruir primeiro
    if (m_buffer != VK_NULL_HANDLE && m_allocator != VK_NULL_HANDLE) {
        vmaDestroyBuffer(m_allocator, m_buffer, m_allocation);
    }

    m_allocator = allocator;
    m_size = size;
    m_type = type;

    // ─────────────────────────────────────────────────────────────────────────
    // Criar Info de Buffer
    // ─────────────────────────────────────────────────────────────────────────

    VkBufferCreateInfo bufferInfo{};
    bufferInfo.sType = VK_STRUCTURE_TYPE_BUFFER_CREATE_INFO;
    bufferInfo.size = size;
    bufferInfo.usage = GetBufferUsageFlags(type);
    bufferInfo.sharingMode = VK_SHARING_MODE_EXCLUSIVE;

    // ─────────────────────────────────────────────────────────────────────────
    // Criar Info de Alocação VMA
    // ─────────────────────────────────────────────────────────────────────────

    VmaAllocationCreateInfo allocInfo{};
    allocInfo.flags = GetAllocationFlags(type);
    allocInfo.usage = VMA_MEMORY_USAGE_AUTO;
    allocInfo.requiredFlags = GetMemoryPropertyFlags(type);

    // ─────────────────────────────────────────────────────────────────────────
    // Alocar Buffer
    // ─────────────────────────────────────────────────────────────────────────

    VkResult result = vmaCreateBuffer(allocator,
                                      &bufferInfo,
                                      &allocInfo,
                                      &m_buffer,
                                      &m_allocation,
                                      nullptr  // pAllocationInfo (opcional)
    );

    if (result != VK_SUCCESS) {
        VK_LOG_ERROR("Falha ao alocar buffer: " << static_cast<int>(result));
        m_buffer = VK_NULL_HANDLE;
        m_allocation = VK_NULL_HANDLE;
        return;
    }

    // Log de sucesso
    const char* typeStr = "";
    switch (type) {
        case BufferType::Vertex:
            typeStr = "Vertex";
            break;
        case BufferType::Index:
            typeStr = "Index";
            break;
        case BufferType::Uniform:
            typeStr = "Uniform";
            break;
        case BufferType::Staging:
            typeStr = "Staging";
            break;
    }

    VK_LOG_INFO("Buffer " << typeStr << " alocado: " << size << " bytes");
}

void Buffer::SetData(const void* data, VkDeviceSize size)
{
    assert(m_allocator != VK_NULL_HANDLE && "Buffer não foi criado");
    assert(data != nullptr && "Dados não podem ser nulos");
    assert(size <= m_size && "Dados são maiores que o buffer");
    assert(m_type == BufferType::Staging ||
           m_type == BufferType::Uniform && "SetData() só funciona com Staging ou Uniform buffers");

    void* mappedPtr = Map();
    if (mappedPtr != nullptr) {
        std::memcpy(mappedPtr, data, size);
        Unmap();
    }
}

void* Buffer::Map()
{
    assert(m_allocator != VK_NULL_HANDLE && "Buffer não foi criado");
    assert(m_type == BufferType::Staging ||
           m_type == BufferType::Uniform && "Map() só funciona com Staging ou Uniform buffers");

    void* mappedPtr = nullptr;
    VkResult result = vmaMapMemory(m_allocator, m_allocation, &mappedPtr);

    if (result != VK_SUCCESS) {
        VK_LOG_ERROR("Falha ao mapear buffer: " << static_cast<int>(result));
        return nullptr;
    }

    return mappedPtr;
}

void Buffer::Unmap()
{
    assert(m_allocator != VK_NULL_HANDLE && "Buffer não foi criado");
    assert(m_type == BufferType::Staging ||
           m_type == BufferType::Uniform && "Unmap() só funciona com Staging ou Uniform buffers");

    vmaUnmapMemory(m_allocator, m_allocation);
}

}  // namespace vklib
