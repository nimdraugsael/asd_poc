class SaveSubdomains
  include Faktory::Job

  faktory_options retry: 5, custom: { unique_for: 10.minutes.to_i }
  # queue_as :ruby_critical

  def perform(domain_id, subdomains)
    domain = Domain.find(domain_id)
    return unless domain.updating_subdomains?

    ActiveRecord::Base.transaction do
      domain.subdomains.destroy_all
      subdomains.each do |subdomain|
        next if domain.domain == subdomain
        domain.subdomains.create(parent_id: domain_id, domain: subdomain)
      end
    end
    domain.finish_subdomain_update!
  end
end
