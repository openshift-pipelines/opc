# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Opc < Formula
  desc "A CLI for OpenShift Pipeline"
  homepage "https://github.com/openshift-pipelines/opc"
  version "1.15.0"

  on_macos do
    on_intel do
      url "https://github.com/openshift-pipelines/opc/releases/download/v1.15.0/opc_1.15.0_darwin_x86_64.tar.gz"
      sha256 "8b5f49a418c6deac1ee42460afc1b2cd7e9872eb9ff337ca23d8617454858f99"

      def install
        bin.install "opc" => "opc"
        output = Utils.popen_read("SHELL=bash #{bin}/opc completion bash")
        (bash_completion/"opc").write output
        output = Utils.popen_read("SHELL=zsh #{bin}/opc completion zsh")
        (zsh_completion/"_opc").write output
        prefix.install_metafiles
      end
    end
    on_arm do
      url "https://github.com/openshift-pipelines/opc/releases/download/v1.15.0/opc_1.15.0_darwin_arm64.tar.gz"
      sha256 "9782c0689560aa0be70667556638144f4205a862fef63c3a633410ee685d23a2"

      def install
        bin.install "opc" => "opc"
        output = Utils.popen_read("SHELL=bash #{bin}/opc completion bash")
        (bash_completion/"opc").write output
        output = Utils.popen_read("SHELL=zsh #{bin}/opc completion zsh")
        (zsh_completion/"_opc").write output
        prefix.install_metafiles
      end
    end
  end

  on_linux do
    on_intel do
      if Hardware::CPU.is_64_bit?
        url "https://github.com/openshift-pipelines/opc/releases/download/v1.15.0/opc_1.15.0_linux_x86_64.tar.gz"
        sha256 "e1d9c64582b3d552d0aef78b8a3f2a48fe1b71e9e67ae57f020e6c8364eb6358"

        def install
          bin.install "opc" => "opc"
          output = Utils.popen_read("SHELL=bash #{bin}/opc completion bash")
          (bash_completion/"opc").write output
          output = Utils.popen_read("SHELL=zsh #{bin}/opc completion zsh")
          (zsh_completion/"_opc").write output
          prefix.install_metafiles
        end
      end
    end
    on_arm do
      if Hardware::CPU.is_64_bit?
        url "https://github.com/openshift-pipelines/opc/releases/download/v1.15.0/opc_1.15.0_linux_arm64.tar.gz"
        sha256 "dbc048a9dd102d18b0b170383da953a9e8b0df2702d32bf4a9119efbb8369c06"

        def install
          bin.install "opc" => "opc"
          output = Utils.popen_read("SHELL=bash #{bin}/opc completion bash")
          (bash_completion/"opc").write output
          output = Utils.popen_read("SHELL=zsh #{bin}/opc completion zsh")
          (zsh_completion/"_opc").write output
          prefix.install_metafiles
        end
      end
    end
  end
end
