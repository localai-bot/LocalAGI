import React from 'react';
import FilterForm from '../FilterForm';

/**
 * FiltersSection component for the agent form
 */
const FiltersSection = ({ formData, setFormData, metadata }) => {
  // Handle filter change
  const handleFilterChange = (index, updatedFilter) => {
    const updatedFilters = [...(formData.filters || [])];
    updatedFilters[index] = updatedFilter;
    setFormData({
      ...formData,
      filters: updatedFilters
    });
  };

  // Handle filter removal
  const handleFilterRemove = (index) => {
    const updatedFilters = [...(formData.filters || [])].filter((_, i) => i !== index);
    setFormData({
      ...formData,
      filters: updatedFilters
    });
  };

  // Handle adding an filter
  const handleAddFilter = () => {
    setFormData({
      ...formData,
      filters: [
        ...(formData.filters || []),
        { name: '', config: '{}' }
      ]
    });
  };

  return (
    <div className="filters-section">
      <h3>Filters</h3>
      <p className="text-muted">
        Configure filters that the agent can perform.
      </p>

      <FilterForm
        filters={formData.filters || []}
        onChange={handleFilterChange}
        onRemove={handleFilterRemove}
        onAdd={handleAddFilter}
        fieldGroups={metadata?.filters || []}
      />
    </div>
  );
};

export default FiltersSection;
