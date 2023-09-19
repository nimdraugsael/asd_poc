class CreateDomains < ActiveRecord::Migration[7.0]
  def change
    create_table :domains do |t|
      t.integer :parent_id
      t.string :domain
      t.datetime :last_crawled_at

      t.timestamps
    end

    add_index :domains, :domain, unique: true
  end
end
