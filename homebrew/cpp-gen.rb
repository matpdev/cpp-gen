# frozen_string_literal: true

# ==============================================================================
# homebrew/cpp-gen.rb — cpp-gen Homebrew formula (TEMPLATE)
# ==============================================================================
# This formula is a reference template. In CI, goreleaser generates and pushes
# the real formula to the matpdev/homebrew-tap repository automatically.
#
# For manual publishing:
#   ./scripts/homebrew-publish.sh
#
# To install from the tap:
#   brew tap matpdev/tap
#   brew install cpp-gen
# ==============================================================================

class CppGen < Formula
  desc "Modern C++ project generator with CMake, package managers, IDE configurations and development tools"
  homepage "https://github.com/matpdev/cpp-gen"
  version "0.1.0"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/matpdev/cpp-gen/releases/download/v#{version}/cpp-gen_#{version}_darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_AMD64"
    end

    on_arm do
      url "https://github.com/matpdev/cpp-gen/releases/download/v#{version}/cpp-gen_#{version}_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_ARM64"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/matpdev/cpp-gen/releases/download/v#{version}/cpp-gen_#{version}_linux_amd64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_AMD64"
    end

    on_arm do
      if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
        url "https://github.com/matpdev/cpp-gen/releases/download/v#{version}/cpp-gen_#{version}_linux_arm64.tar.gz"
        sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"
      end
    end
  end

  def install
    bin.install "cpp-gen"
  end

  test do
    system "#{bin}/cpp-gen", "--version"
  end
end
