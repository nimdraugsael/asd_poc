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
    priority = freshly_created? ? "critical" : "normal"
    subject = "jobs.go.#{priority}.1"

    payload = {
      job: "EnumerateSubdomains",
      params: {
        domain: domain,
      }
    }
    $nats.publish(subject, payload.to_json)
  end
end
