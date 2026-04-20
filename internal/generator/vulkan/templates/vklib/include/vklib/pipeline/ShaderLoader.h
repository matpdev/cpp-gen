#pragma once

/**
 * @file ShaderLoader.h
 * @brief Utilitário para carregar SPIR-V e criar VkShaderModule.
 *
 * Uso com RAII (recomendado — módulo é destruído ao sair do escopo):
 * @code
 *   vklib::ShaderModule vert(device, SHADER_DIR "triangle.vert.spv");
 *   vklib::ShaderModule frag(device, SHADER_DIR "triangle.frag.spv");
 *
 *   auto pipeline = PipelineBuilder()
 *       .AddShader(VK_SHADER_STAGE_VERTEX_BIT,   vert, "main")
 *       .AddShader(VK_SHADER_STAGE_FRAGMENT_BIT, frag, "main")
 *       .Build();
 *   // vert e frag são destruídos aqui — pipeline já foi criada, tudo certo
 * @endcode
 *
 * Uso estático (gerencia lifetime manualmente):
 * @code
 *   VkShaderModule mod = ShaderLoader::LoadModule(device, "triangle.vert.spv");
 *   // ... usar ...
 *   vkDestroyShaderModule(device, mod, nullptr);
 * @endcode
 */

#include <vulkan/vulkan.h>

#include <cstdint>
#include <filesystem>
#include <vector>

namespace vklib {

// =============================================================================
// ShaderLoader — métodos estáticos
// =============================================================================

class ShaderLoader
{
public:
    /**
     * @brief Carrega um arquivo SPIR-V (.spv) do disco
     * @param path Caminho para o arquivo .spv
     * @return Palavras SPIR-V em vetor de uint32_t
     * @throws std::runtime_error se o arquivo não existir ou for inválido
     */
    static std::vector<uint32_t> LoadSpirV(const std::filesystem::path& path);

    /**
     * @brief Cria VkShaderModule a partir de dados SPIR-V
     * @param device Dispositivo Vulkan
     * @param spirv  Dados SPIR-V (vetor de uint32_t)
     * @return VkShaderModule — destruir com vkDestroyShaderModule após a pipeline ser criada
     */
    static VkShaderModule CreateModule(VkDevice device, const std::vector<uint32_t>& spirv);

    /**
     * @brief Atalho: carrega .spv e cria VkShaderModule em uma chamada
     * @param device Dispositivo Vulkan
     * @param path   Caminho para o arquivo .spv
     * @return VkShaderModule — destruir com vkDestroyShaderModule após a pipeline ser criada
     */
    static VkShaderModule LoadModule(VkDevice device, const std::filesystem::path& path);
};

// =============================================================================
// ShaderModule — RAII wrapper para VkShaderModule
// =============================================================================

/**
 * @brief Wrapper RAII que destrói automaticamente o VkShaderModule ao sair do escopo.
 *
 * VkShaderModule pode ser destruído logo após a criação da pipeline.
 * Este wrapper garante isso sem que o usuário precise se lembrar de chamar
 * vkDestroyShaderModule manualmente.
 */
class ShaderModule
{
public:
    ShaderModule() = default;

    /** @brief Carrega .spv e cria o módulo */
    ShaderModule(VkDevice device, const std::filesystem::path& path);

    /** @brief Cria o módulo a partir de SPIR-V já carregado */
    ShaderModule(VkDevice device, const std::vector<uint32_t>& spirv);

    ~ShaderModule();

    // Não copiável
    ShaderModule(const ShaderModule&) = delete;
    ShaderModule& operator=(const ShaderModule&) = delete;

    // Movível
    ShaderModule(ShaderModule&& other) noexcept;
    ShaderModule& operator=(ShaderModule&& other) noexcept;

    /** @brief Handle Vulkan subjacente */
    VkShaderModule Get() const
    {
        return m_module;
    }

    /** @brief Converte implicitamente para VkShaderModule (uso direto no builder) */
    operator VkShaderModule() const
    {
        return m_module;
    }

    /** @brief true se o módulo foi criado com sucesso */
    bool IsValid() const
    {
        return m_module != VK_NULL_HANDLE;
    }

private:
    VkDevice m_device = VK_NULL_HANDLE;
    VkShaderModule m_module = VK_NULL_HANDLE;
};

}  // namespace vklib
