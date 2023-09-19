class Domain < ApplicationRecord
  include AASM

  has_many :subdomains, foreign_key: :parent_id, class_name: "Domain", dependent: :destroy

  aasm do
    state :freshly_created, initial: true
    state :updating_subdomains, :updated_subdomains

    event :start_subdomain_update do
      transitions from: :freshly_created, to: :updating_subdomains, after: :aasm_start_subdomain_enumeration
      transitions from: :updated_subdomains, to: :updating_subdomains, after: :aasm_start_subdomain_enumeration
    end

    event :finish_subdomain_update do
      transitions from: :updating_subdomains, to: :updated_subdomains
    end
  end

  def aasm_start_subdomain_enumeration
    queue = freshly_created? ? "go_critical" : "go_default"

    payload = {
      jobtype: "EnumerateSubdomains",
      queue: queue,
      retry: 10, 
      args: [self.id, domain], 
    }

    Faktory::Client.new.push(payload)
  end
end
