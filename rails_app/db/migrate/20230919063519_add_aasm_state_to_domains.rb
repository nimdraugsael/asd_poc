class AddAasmStateToDomains < ActiveRecord::Migration[7.0]
  def change
    add_column :domains, :aasm_state, :string
  end
end
