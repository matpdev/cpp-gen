// GLEW precisa ser incluído antes de qualquer header que puxe <GL/gl.h>,
// incluindo o próprio GLFW — por isso a ordem dos includes abaixo importa.
#include <GL/glew.h>
#include <GLFW/glfw3.h>

#include <cstdio>
#include <cstdlib>

namespace {

constexpr int kWindowWidth = 800;
constexpr int kWindowHeight = 600;

const char* kVertexShaderSource = R"glsl(
#version 330 core
layout (location = 0) in vec3 aPos;

void main() {
    gl_Position = vec4(aPos, 1.0);
}
)glsl";

const char* kFragmentShaderSource = R"glsl(
#version 330 core
out vec4 FragColor;

void main() {
    FragColor = vec4(1.0, 0.5, 0.2, 1.0);
}
)glsl";

void glfwErrorCallback(int error, const char* description) {
    std::fprintf(stderr, "Erro GLFW [%d]: %s\n", error, description);
}

GLuint compileShader(GLenum type, const char* source) {
    GLuint shader = glCreateShader(type);
    glShaderSource(shader, 1, &source, nullptr);
    glCompileShader(shader);

    GLint success = 0;
    glGetShaderiv(shader, GL_COMPILE_STATUS, &success);
    if (!success) {
        char log[512];
        glGetShaderInfoLog(shader, sizeof(log), nullptr, log);
        std::fprintf(stderr, "Erro ao compilar shader: %s\n", log);
        std::exit(EXIT_FAILURE);
    }
    return shader;
}

GLuint linkProgram(GLuint vertexShader, GLuint fragmentShader) {
    GLuint program = glCreateProgram();
    glAttachShader(program, vertexShader);
    glAttachShader(program, fragmentShader);
    glLinkProgram(program);

    GLint success = 0;
    glGetProgramiv(program, GL_LINK_STATUS, &success);
    if (!success) {
        char log[512];
        glGetProgramInfoLog(program, sizeof(log), nullptr, log);
        std::fprintf(stderr, "Erro ao linkar shader program: %s\n", log);
        std::exit(EXIT_FAILURE);
    }
    return program;
}

}  // namespace

int main() {
    glfwSetErrorCallback(glfwErrorCallback);
    if (!glfwInit()) {
        std::fprintf(stderr, "Falha ao inicializar o GLFW\n");
        return EXIT_FAILURE;
    }

    glfwWindowHint(GLFW_CONTEXT_VERSION_MAJOR, 3);
    glfwWindowHint(GLFW_CONTEXT_VERSION_MINOR, 3);
    glfwWindowHint(GLFW_OPENGL_PROFILE, GLFW_OPENGL_CORE_PROFILE);
#ifdef __APPLE__
    glfwWindowHint(GLFW_OPENGL_FORWARD_COMPAT, GLFW_TRUE);
#endif

    GLFWwindow* window = glfwCreateWindow(kWindowWidth, kWindowHeight, "{{.Name}}", nullptr, nullptr);
    if (!window) {
        std::fprintf(stderr, "Falha ao criar a janela GLFW\n");
        glfwTerminate();
        return EXIT_FAILURE;
    }

    glfwMakeContextCurrent(window);
    glfwSwapInterval(1);  // vsync

    glewExperimental = GL_TRUE;
    if (glewInit() != GLEW_OK) {
        std::fprintf(stderr, "Falha ao inicializar o GLEW\n");
        glfwTerminate();
        return EXIT_FAILURE;
    }

    GLuint vertexShader = compileShader(GL_VERTEX_SHADER, kVertexShaderSource);
    GLuint fragmentShader = compileShader(GL_FRAGMENT_SHADER, kFragmentShaderSource);
    GLuint shaderProgram = linkProgram(vertexShader, fragmentShader);
    glDeleteShader(vertexShader);
    glDeleteShader(fragmentShader);

    constexpr float vertices[] = {
        -0.5f, -0.5f, 0.0f,
         0.5f, -0.5f, 0.0f,
         0.0f,  0.5f, 0.0f,
    };

    GLuint vao = 0;
    GLuint vbo = 0;
    glGenVertexArrays(1, &vao);
    glGenBuffers(1, &vbo);

    glBindVertexArray(vao);
    glBindBuffer(GL_ARRAY_BUFFER, vbo);
    glBufferData(GL_ARRAY_BUFFER, sizeof(vertices), vertices, GL_STATIC_DRAW);
    glVertexAttribPointer(0, 3, GL_FLOAT, GL_FALSE, 3 * sizeof(float), nullptr);
    glEnableVertexAttribArray(0);
    glBindBuffer(GL_ARRAY_BUFFER, 0);
    glBindVertexArray(0);

    while (!glfwWindowShouldClose(window)) {
        glClearColor(0.1F, 0.1F, 0.12F, 1.0F);
        glClear(GL_COLOR_BUFFER_BIT);

        glUseProgram(shaderProgram);
        glBindVertexArray(vao);
        glDrawArrays(GL_TRIANGLES, 0, 3);

        glfwSwapBuffers(window);
        glfwPollEvents();
    }

    glDeleteVertexArrays(1, &vao);
    glDeleteBuffers(1, &vbo);
    glDeleteProgram(shaderProgram);

    glfwTerminate();
    return EXIT_SUCCESS;
}
