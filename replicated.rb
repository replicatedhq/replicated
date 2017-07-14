class Replicated < Formula
  desc "Manage your app's channels and releases from the command line"
  homepage "https://replicated.com"
  url "https://github.com/replicatedhq/replicated/releases/download/v0.1.1/replicated_0.1.1_darwin_amd64.tar.gz"
  version "0.1.1"
  sha256 "ccda0dbf9da532fbfb339cc4fbfd2c113adda00f58cb712a139482a305e9be51"

  def install
    bin.install "replicated"
  end
end
