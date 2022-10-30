require 'subprocess/shell'

RSpec.describe Subprocess::Shell do
  let(:open3) { class_double(Open3) }

  before do
    stub_const 'Open3', open3
  end

  context 'with #exec' do
    before do
      allow(open3).to receive(:capture2e).and_return([' OK ', 0])
    end

    it 'returns the result' do
      expect(described_class.exec('test_command')).to eq ['OK', 0]
    end
  end

  context 'with backtick commands' do
    it 'overwrites Kernel.<backtick>' do
      expect { `which ruby` }.to raise_error(SecurityError)
    end
  end
end
