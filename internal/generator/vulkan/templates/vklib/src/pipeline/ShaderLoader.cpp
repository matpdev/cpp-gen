#include <vklib/common/Defines.h>
#include <vklib/pipeline/ShaderLoader.h>

#include <fstream>
#include <stdexcept>
#include <utility>

namespace vklib {

// =============================================================================
// ShaderLoader
// =============================================================================

std::vector<uint32_t> ShaderLoader::LoadSpirV(const std::filesystem::path& path)
{
    std::ifstream file(path, std::ios::binary | std::ios::ate);
    if (!file.is_open()) {
        throw std::runtime_error("ShaderLoader: arquivo não encontrado: " + path.string() +
                                 "\n"
                                 "  Verifique se os shaders foram compilados (cmake --build .)");
    }

    const auto fileSize = static_cast<std::streamsize>(file.tellg());

    if (fileSize == 0) {
        throw std::runtime_error("ShaderLoader: arquivo vazio: " + path.string());
    }
    if (fileSize % static_cast<std::streamsize>(sizeof(uint32_t)) != 0) {
        throw std::runtime_error(
            "ShaderLoader: SPIR-V inválido — tamanho não é múltiplo de 4 bytes: " + path.string());
    }

    std::vector<uint32_t> buffer(static_cast<size_t>(fileSize) / sizeof(uint32_t));

    file.seekg(0);
    file.read(reinterpret_cast<char*>(buffer.data()), fileSize);

    if (!file) {
        throw std::runtime_error("ShaderLoader: falha ao ler: " + path.string());
    }

    VK_LOG_INFO("SPIR-V carregado: " << path.filename().string() << "  (" << fileSize << " bytes)");
    return buffer;
}

VkShaderModule ShaderLoader::CreateModule(VkDevice device, const std::vector<uint32_t>& spirv)
{
    VkShaderModuleCreateInfo createInfo{};
    createInfo.sType = VK_STRUCTURE_TYPE_SHADER_MODULE_CREATE_INFO;
    createInfo.codeSize = spirv.size() * sizeof(uint32_t);
    createInfo.pCode = spirv.data();

    VkShaderModule shaderModule = VK_NULL_HANDLE;
    VK_CHECK(vkCreateShaderModule(device, &createInfo, nullptr, &shaderModule));

    return shaderModule;
}

VkShaderModule ShaderLoader::LoadModule(VkDevice device, const std::filesystem::path& path)
{
    return CreateModule(device, LoadSpirV(path));
}

// =============================================================================
// ShaderModule (RAII)
// =============================================================================

ShaderModule::ShaderModule(VkDevice device, const std::filesystem::path& path)
    : m_device(device), m_module(ShaderLoader::LoadModule(device, path))
{}

ShaderModule::ShaderModule(VkDevice device, const std::vector<uint32_t>& spirv)
    : m_device(device), m_module(ShaderLoader::CreateModule(device, spirv))
{}

ShaderModule::~ShaderModule()
{
    if (m_module != VK_NULL_HANDLE && m_device != VK_NULL_HANDLE) {
        vkDestroyShaderModule(m_device, m_module, nullptr);
        m_module = VK_NULL_HANDLE;
    }
}

ShaderModule::ShaderModule(ShaderModule&& other) noexcept
    : m_device(std::exchange(other.m_device, VK_NULL_HANDLE)),
      m_module(std::exchange(other.m_module, VK_NULL_HANDLE))
{}

ShaderModule& ShaderModule::operator=(ShaderModule&& other) noexcept
{
    if (this != &other) {
        if (m_module != VK_NULL_HANDLE && m_device != VK_NULL_HANDLE)
            vkDestroyShaderModule(m_device, m_module, nullptr);

        m_device = std::exchange(other.m_device, VK_NULL_HANDLE);
        m_module = std::exchange(other.m_module, VK_NULL_HANDLE);
    }
    return *this;
}

}  // namespace vklib
